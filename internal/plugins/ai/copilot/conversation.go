package copilot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// ConversationManager handles Microsoft Copilot conversation lifecycle
type ConversationManager struct {
	activeConversations map[string]string // Fabric session ID -> Copilot conversation ID
	mutex               *sync.RWMutex
	httpClient          *http.Client
	baseURL             string
	accessToken         string
}

// NewConversationManager creates a new conversation manager
func NewConversationManager(baseURL, accessToken string) *ConversationManager {
	return &ConversationManager{
		activeConversations: make(map[string]string),
		mutex:               &sync.RWMutex{},
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		baseURL:     baseURL,
		accessToken: accessToken,
	}
}

// GetOrCreateConversation gets an existing conversation or creates a new one
func (cm *ConversationManager) GetOrCreateConversation(sessionID string) (string, error) {
	cm.mutex.RLock()
	if convID, exists := cm.activeConversations[sessionID]; exists {
		cm.mutex.RUnlock()
		return convID, nil
	}
	cm.mutex.RUnlock()

	// Create new conversation
	convID, err := cm.createConversation()
	if err != nil {
		return "", fmt.Errorf("failed to create conversation: %w", err)
	}

	// Store the conversation ID
	cm.mutex.Lock()
	cm.activeConversations[sessionID] = convID
	cm.mutex.Unlock()

	return convID, nil
}

// createConversation creates a new Microsoft Copilot conversation
func (cm *ConversationManager) createConversation() (string, error) {
	req, err := http.NewRequestWithContext(context.Background(), "POST", cm.baseURL+"/conversations", nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+cm.accessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Empty JSON body as per API spec
	req.Body = http.NoBody

	resp, err := cm.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("failed to create conversation, status: %d", resp.StatusCode)
	}

	var convResponse ConversationResponse
	if err := json.NewDecoder(resp.Body).Decode(&convResponse); err != nil {
		return "", err
	}

	return convResponse.ID, nil
}

// SendMessage sends a message to a conversation
func (cm *ConversationManager) SendMessage(conversationID, message string) (*ConversationResponse, error) {
	// Build request payload
	payload := ChatRequest{
		Message: CopilotMessage{
			Text: message,
		},
		LocationHint: LocationHint{
			TimeZone: "UTC", // Default to UTC, could be configurable
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/conversations/%s/chat", cm.baseURL, conversationID)
	req, err := http.NewRequestWithContext(context.Background(), "POST", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+cm.accessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Set request body properly
	if payloadBytes != nil {
		req.Body = http.NoBody // Will be set below
	}

	// Create request with body
	req, err = http.NewRequestWithContext(context.Background(), "POST", url, bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+cm.accessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := cm.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to send message, status: %d", resp.StatusCode)
	}

	var convResponse ConversationResponse
	if err := json.NewDecoder(resp.Body).Decode(&convResponse); err != nil {
		return nil, err
	}

	return &convResponse, nil
}

// CleanupConversation removes a conversation from the active list
func (cm *ConversationManager) CleanupConversation(sessionID string) {
	cm.mutex.Lock()
	delete(cm.activeConversations, sessionID)
	cm.mutex.Unlock()
}

// CleanupOldConversations removes conversations older than specified duration
func (cm *ConversationManager) CleanupOldConversations(maxAge time.Duration) {
	// TODO: Implement conversation age tracking and cleanup
	// This would require storing creation timestamps
}
