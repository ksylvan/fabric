package copilot

import (
	"testing"

	"github.com/danielmiessler/fabric/internal/chat"
)

func TestNewClient(t *testing.T) {
	client := NewClient()

	if client.GetName() != "Microsoft Copilot" {
		t.Errorf("Expected name 'Microsoft Copilot', got '%s'", client.GetName())
	}

	models, err := client.ListModels()
	if err != nil {
		t.Fatalf("Failed to list models: %v", err)
	}

	if len(models) != 1 || models[0] != "copilot-enterprise" {
		t.Errorf("Expected models ['copilot-enterprise'], got %v", models)
	}

	if client.NeedsRawMode("copilot-enterprise") {
		t.Error("Copilot should not need raw mode")
	}
}

func TestConvertFabricMessages(t *testing.T) {
	client := NewClient()

	msgs := []*chat.ChatCompletionMessage{
		{Role: chat.ChatMessageRoleSystem, Content: "You are a helpful assistant."},
		{Role: chat.ChatMessageRoleUser, Content: "Hello, how are you?"},
	}

	result := client.convertFabricMessages(msgs)
	expected := "System: You are a helpful assistant.\n\nUser: Hello, how are you?\n\n"

	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestExtractResponseText(t *testing.T) {
	client := NewClient()

	// Test with nil response
	result := client.extractResponseText(nil)
	if result != "" {
		t.Errorf("Expected empty string for nil response, got '%s'", result)
	}

	// Test with empty messages
	emptyResponse := &ConversationResponse{Messages: []CopilotConversationMessage{}}
	result = client.extractResponseText(emptyResponse)
	if result != "" {
		t.Errorf("Expected empty string for empty messages, got '%s'", result)
	}

	// Test with valid response
	validResponse := &ConversationResponse{
		Messages: []CopilotConversationMessage{
			{Text: "User: Hello"},
			{Text: "Hello! How can I help you today?"},
		},
	}
	result = client.extractResponseText(validResponse)
	expected := "Hello! How can I help you today?"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestConversationManager(t *testing.T) {
	// Test conversation manager creation
	cm := NewConversationManager("https://test.com", "test-token")

	if cm.baseURL != "https://test.com" {
		t.Errorf("Expected base URL 'https://test.com', got '%s'", cm.baseURL)
	}

	if cm.accessToken != "test-token" {
		t.Errorf("Expected access token 'test-token', got '%s'", cm.accessToken)
	}
}

func TestIsConfigured(t *testing.T) {
	client := NewClient()

	// Test with no configuration
	if client.IsConfigured() {
		t.Error("Client should not be configured without credentials")
	}

	// Test with client ID only
	client.ClientID.Value = "test-client-id"
	if client.IsConfigured() {
		t.Error("Client should not be configured with only client ID")
	}

	// Test with client ID and secret (app-only mode)
	client.ClientSecret.Value = "test-client-secret"
	if !client.IsConfigured() {
		t.Error("Client should be configured with client ID and secret")
	}
}
