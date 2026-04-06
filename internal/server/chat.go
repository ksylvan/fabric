package restapi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/danielmiessler/fabric/internal/chat"

	"github.com/danielmiessler/fabric/internal/core"
	"github.com/danielmiessler/fabric/internal/domain"
	"github.com/danielmiessler/fabric/internal/i18n"
	debuglog "github.com/danielmiessler/fabric/internal/log"
	"github.com/danielmiessler/fabric/internal/plugins/db/fsdb"
	"github.com/gin-gonic/gin"
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
	Language           string          `json:"language"`
	ModelContextLength int             `json:"modelContextLength,omitempty"` // Context window size
	domain.ChatOptions                 // Embed the ChatOptions from common package
}

type StreamResponse struct {
	Type    string                `json:"type"`             // "content", "usage", "error", "complete"
	Format  string                `json:"format,omitempty"` // "markdown", "mermaid", "plain"
	Content string                `json:"content,omitempty"`
	Usage   *domain.UsageMetadata `json:"usage,omitempty"`
}

const localDevCORSOrigin = "http://localhost:5173"
const chatSSECacheControl = "no-cache, no-store, must-revalidate"
const clientSafeChatErrorMessage = "chat request failed"

func NewChatHandler(r *gin.Engine, registry *core.PluginRegistry, db *fsdb.Db) *ChatHandler {
	handler := &ChatHandler{
		registry: registry,
		db:       db,
	}

	r.POST("/chat", handler.HandleChat)
	r.OPTIONS("/chat", handler.HandleChatOptions)

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
	applyLocalDevCORSHeaders(c)

	if err := decodeStrictJSON(c, &request); err != nil {
		log.Printf("Error binding JSON: %v", err)
		c.Writer.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf(i18n.T("server_invalid_request_format"), err)})
		return
	}

	debuglog.Debug(debuglog.Basic, "Received chat request - language=%q prompts=%d\n", request.Language, len(request.Prompts))

	// Set headers for SSE
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", chatSSECacheControl)
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Pragma", "no-cache")
	c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	clientGone := c.Request.Context().Done()

	for i, prompt := range request.Prompts {
		select {
		case <-clientGone:
			debuglog.Debug(debuglog.Basic, "Client disconnected from /chat stream\n")
			return
		default:
			debuglog.Debug(debuglog.Basic, "Processing prompt %d: model=%s pattern=%s context=%s\n",
				i+1, prompt.Model, prompt.PatternName, prompt.ContextName)

			streamChan := make(chan domain.StreamUpdate, 16)

			go func(p PromptRequest) {
				defer close(streamChan)

				chatter, err := h.registry.GetChatter(
					p.Model,
					request.ModelContextLength,
					p.Vendor,
					true,
					false,
				)
				if err != nil {
					log.Printf("Error creating chatter: %v", err)
					streamChan <- newClientSafeChatErrorUpdate()
					return
				}

				chatReq := buildPromptChatRequest(p, request.Language)

				opts := &domain.ChatOptions{
					Model:            p.Model,
					Temperature:      request.Temperature,
					TopP:             request.TopP,
					FrequencyPenalty: request.FrequencyPenalty,
					PresencePenalty:  request.PresencePenalty,
					Thinking:         request.Thinking,
					Search:           request.Search,
					SearchLocation:   request.SearchLocation,
					UpdateChan:       streamChan,
					Quiet:            true,
				}

				_, err = chatter.Send(c.Request.Context(), chatReq, opts)
				if err != nil {
					log.Printf("Error from chatter.Send: %v", err)
					streamChan <- newClientSafeChatErrorUpdate()
					return
				}
			}(prompt)

			sawError := false
			for update := range streamChan {
				select {
				case <-clientGone:
					return
				default:
					response, ok := streamResponseForUpdate(update)
					if !ok {
						continue
					}
					if response.Type == "error" {
						sawError = true
					}

					if err := writeSSEResponse(c.Writer, response); err != nil {
						log.Printf("Error writing response: %v", err)
						return
					}
				}
			}
			if sawError {
				return
			}

			completeResponse := StreamResponse{
				Type:    "complete",
				Format:  "plain",
				Content: "",
			}
			if err := writeSSEResponse(c.Writer, completeResponse); err != nil {
				log.Printf("Error writing completion response: %v", err)
				return
			}
		}
	}
}

func newClientSafeChatErrorUpdate() domain.StreamUpdate {
	return domain.StreamUpdate{
		Type:    domain.StreamTypeError,
		Content: clientSafeChatErrorMessage,
	}
}

func streamResponseForUpdate(update domain.StreamUpdate) (StreamResponse, bool) {
	switch update.Type {
	case domain.StreamTypeContent:
		return StreamResponse{
			Type:    "content",
			Format:  detectFormat(update.Content),
			Content: update.Content,
		}, true
	case domain.StreamTypeUsage:
		return StreamResponse{
			Type:  "usage",
			Usage: update.Usage,
		}, true
	case domain.StreamTypeError:
		return StreamResponse{
			Type:    "error",
			Format:  "plain",
			Content: clientSafeChatErrorMessage,
		}, true
	default:
		return StreamResponse{}, false
	}
}

func (h *ChatHandler) HandleChatOptions(c *gin.Context) {
	applyLocalDevCORSHeaders(c)
	c.Status(http.StatusNoContent)
}

func buildPromptChatRequest(p PromptRequest, language string) *domain.ChatRequest {
	return &domain.ChatRequest{
		Message: &chat.ChatCompletionMessage{
			Role:    chat.ChatMessageRoleUser,
			Content: p.UserInput,
		},
		PatternName:              p.PatternName,
		ContextName:              p.ContextName,
		SessionName:              p.SessionName,
		PatternVariables:         p.Variables,
		RestrictTemplateFeatures: true,
		StrategyName:             p.StrategyName,
		Language:                 language,
	}
}

func applyLocalDevCORSHeaders(c *gin.Context) {
	origin := c.GetHeader("Origin")
	if origin != localDevCORSOrigin {
		return
	}

	headers := c.Writer.Header()
	headers.Set("Access-Control-Allow-Origin", origin)
	headers.Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	headers.Set("Access-Control-Allow-Headers", "Content-Type, Accept, X-API-Key")
	headers.Set("Access-Control-Max-Age", "600")
	headers.Add("Vary", "Origin")
}

func writeSSEResponse(w gin.ResponseWriter, response StreamResponse) error {
	data, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("%s", fmt.Sprintf(i18n.T("server_error_marshaling_response"), err))
	}

	if _, err := fmt.Fprintf(w, "data: %s\n\n", string(data)); err != nil {
		return fmt.Errorf("%s", fmt.Sprintf(i18n.T("server_error_writing_response"), err))
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
