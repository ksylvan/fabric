package restapi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/danielmiessler/fabric/internal/chat"

	"github.com/danielmiessler/fabric/internal/core"
	"github.com/danielmiessler/fabric/internal/domain"
	"github.com/danielmiessler/fabric/internal/plugins/db/fsdb"
	"github.com/gin-gonic/gin"
)

const (
	// frontendDevServerURL is the default URL for the frontend development server (Vite default port)
	frontendDevServerURL = "http://localhost:5173"
)

type ChatHandler struct {
	registry *core.PluginRegistry
	db       *fsdb.Db
}

type PromptRequest struct {
	UserInput    string            `json:"userInput"`
	Vendor       string            `json:"vendor"`
	Model        string            `json:"model"`
	ContextName  string            `json:"contextName"`
	PatternName  string            `json:"patternName"`
	StrategyName string            `json:"strategyName"`        // Optional strategy name
	SessionName  string            `json:"sessionName"`         // Session name for multi-turn conversations
	Variables    map[string]string `json:"variables,omitempty"` // Pattern variables
}

type ChatRequest struct {
	Prompts            []PromptRequest `json:"prompts"`
	Language           string          `json:"language"` // Add Language field to bind from request
	domain.ChatOptions                 // Embed the ChatOptions from common package
}

type StreamResponse struct {
	Type    string `json:"type"`    // "content", "error", "complete"
	Format  string `json:"format"`  // "markdown", "mermaid", "plain"
	Content string `json:"content"` // The actual content
}

func NewChatHandler(r *gin.Engine, registry *core.PluginRegistry, db *fsdb.Db) *ChatHandler {
	handler := &ChatHandler{
		registry: registry,
		db:       db,
	}

	r.POST("/chat", handler.HandleChat)

	return handler
}

// HandleChat godoc
// @Summary Stream chat completions
// @Description Stream AI responses using Server-Sent Events (SSE)
// @Tags chat
// @Accept json
// @Produce text/event-stream
// @Param request body ChatRequest true "Chat request with prompts and options"
// @Success 200 {object} StreamResponse "Streaming response"
// @Failure 400 {object} map[string]string
// @Security ApiKeyAuth
// @Router /chat [post]
func (h *ChatHandler) HandleChat(c *gin.Context) {
	var request ChatRequest

	if err := c.BindJSON(&request); err != nil {
		h.handleBindError(c, err)
		return
	}

	log.Printf("Received chat request - Language: '%s', Prompts: %d", request.Language, len(request.Prompts))

	h.setupSSEHeaders(c)
	clientGone := c.Writer.CloseNotify()

	for i, prompt := range request.Prompts {
		if h.shouldStopProcessing(clientGone) {
			return
		}

		log.Printf("Processing prompt %d: Model=%s Pattern=%s Context=%s",
			i+1, prompt.Model, prompt.PatternName, prompt.ContextName)

		streamChan := make(chan string)
		go h.processPrompt(prompt, request, streamChan)

		if err := h.streamResponses(c, streamChan, clientGone); err != nil {
			return
		}
	}
}

func (h *ChatHandler) handleBindError(c *gin.Context, err error) {
	log.Printf("Error binding JSON: %v", err)
	c.Writer.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
	c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request format: %v", err)})
}

func (h *ChatHandler) setupSSEHeaders(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/readystream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Access-Control-Allow-Origin", frontendDevServerURL)
	c.Writer.Header().Set("X-Accel-Buffering", "no")
}

func (h *ChatHandler) shouldStopProcessing(clientGone <-chan bool) bool {
	select {
	case <-clientGone:
		log.Printf("Client disconnected")
		return true
	default:
		return false
	}
}

func (h *ChatHandler) processPrompt(p PromptRequest, request ChatRequest, streamChan chan string) {
	defer close(streamChan)

	userInput := h.loadStrategyPrompt(p)

	chatter, err := h.registry.GetChatter(p.Model, 2048, p.Vendor, "", false, false)
	if err != nil {
		log.Printf("Error creating chatter: %v", err)
		streamChan <- fmt.Sprintf("Error: %v", err)
		return
	}

	chatReq := &domain.ChatRequest{
		Message: &chat.ChatCompletionMessage{
			Role:    "user",
			Content: userInput,
		},
		PatternName:      p.PatternName,
		ContextName:      p.ContextName,
		SessionName:      p.SessionName,
		PatternVariables: p.Variables,
		Language:         request.Language,
	}

	opts := &domain.ChatOptions{
		Model:            p.Model,
		Temperature:      request.Temperature,
		TopP:             request.TopP,
		FrequencyPenalty: request.FrequencyPenalty,
		PresencePenalty:  request.PresencePenalty,
		Thinking:         request.Thinking,
	}

	h.sendChatRequest(chatReq, opts, chatter, streamChan)
}

func (h *ChatHandler) loadStrategyPrompt(p PromptRequest) string {
	if p.StrategyName == "" {
		return p.UserInput
	}

	strategyFile := filepath.Join(os.Getenv("HOME"), ".config", "fabric", "strategies", p.StrategyName+".json")
	data, err := os.ReadFile(strategyFile)
	if err != nil {
		return p.UserInput
	}

	var s struct {
		Prompt string `json:"prompt"`
	}
	if err := json.Unmarshal(data, &s); err != nil || s.Prompt == "" {
		return p.UserInput
	}

	return s.Prompt + "\n" + p.UserInput
}

func (h *ChatHandler) sendChatRequest(chatReq *domain.ChatRequest, opts *domain.ChatOptions, chatter *core.Chatter, streamChan chan string) {
	session, err := chatter.Send(chatReq, opts)
	if err != nil {
		log.Printf("Error from chatter.Send: %v", err)
		streamChan <- fmt.Sprintf("Error: %v", err)
		return
	}

	if session == nil {
		log.Printf("No session returned from chatter.Send")
		streamChan <- "Error: No response from model"
		return
	}

	lastMsg := session.GetLastMessage()
	if lastMsg != nil {
		streamChan <- lastMsg.Content
	} else {
		log.Printf("No message content in session")
		streamChan <- "Error: No response content"
	}
}

func (h *ChatHandler) streamResponses(c *gin.Context, streamChan chan string, clientGone <-chan bool) error {
	for content := range streamChan {
		if h.shouldStopProcessing(clientGone) {
			return fmt.Errorf("client disconnected")
		}

		response := h.createStreamResponse(content)
		if err := writeSSEResponse(c.Writer, response); err != nil {
			log.Printf("Error writing response: %v", err)
			return err
		}
	}

	return h.sendCompletionResponse(c)
}

func (h *ChatHandler) createStreamResponse(content string) StreamResponse {
	if strings.HasPrefix(content, "Error:") {
		return StreamResponse{
			Type:    "error",
			Format:  "plain",
			Content: content,
		}
	}
	return StreamResponse{
		Type:    "content",
		Format:  detectFormat(content),
		Content: content,
	}
}

func (h *ChatHandler) sendCompletionResponse(c *gin.Context) error {
	completeResponse := StreamResponse{
		Type:    "complete",
		Format:  "plain",
		Content: "",
	}
	if err := writeSSEResponse(c.Writer, completeResponse); err != nil {
		log.Printf("Error writing completion response: %v", err)
		return err
	}
	return nil
}

func writeSSEResponse(w gin.ResponseWriter, response StreamResponse) error {
	data, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("error marshaling response: %w", err)
	}

	if _, err := fmt.Fprintf(w, "data: %s\n\n", string(data)); err != nil {
		return fmt.Errorf("error writing response: %w", err)
	}

	w.(http.Flusher).Flush()
	return nil
}

func detectFormat(content string) string {
	if strings.HasPrefix(content, "graph TD") ||
		strings.HasPrefix(content, "gantt") ||
		strings.HasPrefix(content, "flowchart") ||
		strings.HasPrefix(content, "sequenceDiagram") ||
		strings.HasPrefix(content, "classDiagram") ||
		strings.HasPrefix(content, "stateDiagram") {
		return "mermaid"
	}
	return "markdown"
}
