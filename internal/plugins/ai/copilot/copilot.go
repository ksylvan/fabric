package copilot

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/danielmiessler/fabric/internal/chat"
	"github.com/danielmiessler/fabric/internal/domain"
	"github.com/danielmiessler/fabric/internal/plugins"
)

const (
	defaultBaseURL = "https://graph.microsoft.com/beta/copilot"
	defaultModel   = "copilot-enterprise"
)

// NewClient creates a new Microsoft Copilot client
func NewClient() (ret *Client) {
	ret = &Client{}

	ret.PluginBase = &plugins.PluginBase{
		Name:            "Microsoft Copilot",
		EnvNamePrefix:   plugins.BuildEnvVariablePrefix("Microsoft Copilot"),
		ConfigureCustom: ret.configure,
	}

	// Authentication setup questions
	ret.ClientID = ret.AddSetupQuestion("Client ID", true)
	ret.ClientSecret = ret.AddSetupQuestion("Client Secret", false)
	ret.TenantID = ret.AddSetupQuestion("Tenant ID", false)
	ret.UseOAuth = ret.AddSetupQuestionBool("Use OAuth flow", false)

	// Initialize conversation manager (will be set after authentication)
	ret.conversationManager = nil

	// Set default model
	ret.models = []string{defaultModel}

	return
}

// Client represents the Microsoft Copilot vendor implementation
type Client struct {
	*plugins.PluginBase

	// Authentication
	ClientID     *plugins.SetupQuestion
	ClientSecret *plugins.SetupQuestion
	TenantID     *plugins.SetupQuestion
	UseOAuth     *plugins.SetupQuestion

	// Internal state
	accessToken         string
	conversationManager *ConversationManager
	models              []string
}

// IsConfigured returns true if the client is properly configured
func (c *Client) IsConfigured() bool {
	// Check if we have a valid access token
	if c.accessToken != "" {
		return true
	}

	// Check if we have enough configuration to authenticate
	if c.ClientID.Value == "" {
		return false
	}

	// For app-only access, we need client secret
	if !plugins.ParseBoolElseFalse(c.UseOAuth.Value) && c.ClientSecret.Value == "" {
		return false
	}

	return true
}

// GetName returns the vendor name
func (c *Client) GetName() string {
	return "Microsoft Copilot"
}

// ListModels returns the available models
func (c *Client) ListModels() ([]string, error) {
	return c.models, nil
}

// NeedsRawMode returns false as Copilot handles all modes
func (c *Client) NeedsRawMode(modelName string) bool {
	return false
}

// Send sends a message to Microsoft Copilot and returns the response
func (c *Client) Send(ctx context.Context, msgs []*chat.ChatCompletionMessage, opts *domain.ChatOptions) (string, error) {
	if c.conversationManager == nil {
		return "", fmt.Errorf("conversation manager not initialized - please configure client first")
	}

	// Convert Fabric messages to single Copilot message
	prompt := c.convertFabricMessages(msgs)

	// Generate session ID for this request
	sessionID := fmt.Sprintf("fabric_%d", time.Now().UnixNano())

	// Get or create conversation
	conversationID, err := c.conversationManager.GetOrCreateConversation(sessionID)
	if err != nil {
		return "", fmt.Errorf("failed to get or create conversation: %w", err)
	}

	// Send message to Copilot
	response, err := c.conversationManager.SendMessage(conversationID, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to send message to Copilot: %w", err)
	}

	// Extract response text
	responseText := c.extractResponseText(response)

	// Clean up conversation
	c.conversationManager.CleanupConversation(sessionID)

	return responseText, nil
}

// SendStream sends a message to Microsoft Copilot and streams the response
func (c *Client) SendStream(msgs []*chat.ChatCompletionMessage, opts *domain.ChatOptions, channel chan string) error {
	defer close(channel)

	// For now, use non-streaming and send complete response
	// TODO: Implement true streaming when Microsoft Copilot supports it
	response, err := c.Send(context.Background(), msgs, opts)
	if err != nil {
		return err
	}

	// Send the complete response as one chunk
	channel <- response
	return nil
}

// configure sets up the client with authentication
func (c *Client) configure() error {
	// Try to authenticate based on configuration
	var err error
	if plugins.ParseBoolElseFalse(c.UseOAuth.Value) {
		err = c.authenticateWithOAuth()
	} else {
		err = c.authenticateWithClientSecret()
	}

	if err != nil {
		return err
	}

	// Initialize conversation manager after successful authentication
	c.initializeConversationManager()
	return nil
}

// convertFabricMessages converts Fabric chat messages to a single Copilot prompt
func (c *Client) convertFabricMessages(msgs []*chat.ChatCompletionMessage) string {
	var prompt strings.Builder

	for _, msg := range msgs {
		if msg.Content == "" {
			continue
		}

		// Add role prefix
		switch msg.Role {
		case chat.ChatMessageRoleSystem:
			prompt.WriteString("System: ")
		case chat.ChatMessageRoleUser:
			prompt.WriteString("User: ")
		case chat.ChatMessageRoleAssistant:
			prompt.WriteString("Assistant: ")
		default:
			prompt.WriteString(fmt.Sprintf("%s: ", msg.Role))
		}

		prompt.WriteString(msg.Content)
		prompt.WriteString("\n\n")
	}

	return prompt.String()
}

// extractResponseText extracts the assistant's response from Copilot conversation
func (c *Client) extractResponseText(response *ConversationResponse) string {
	if response == nil || len(response.Messages) == 0 {
		return ""
	}

	// Find the last assistant message
	for i := len(response.Messages) - 1; i >= 0; i-- {
		msg := response.Messages[i]
		if msg.Text != "" {
			// Check if this looks like an assistant response (not echoing user input)
			if !strings.Contains(strings.ToLower(msg.Text), "user:") && !strings.Contains(strings.ToLower(msg.Text), "system:") {
				return msg.Text
			}
		}
	}

	return ""
}

// initializeConversationManager sets up the conversation manager after authentication
func (c *Client) initializeConversationManager() {
	c.conversationManager = NewConversationManager(defaultBaseURL, c.accessToken)
}
