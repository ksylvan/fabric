package copilot

import "time"

// CopilotMessage represents a message in Microsoft Copilot API
type CopilotMessage struct {
	Text string `json:"text"`
}

// LocationHint provides location information for Copilot
type LocationHint struct {
	TimeZone string `json:"timeZone"`
}

// ChatRequest represents a request to Microsoft Copilot chat API
type ChatRequest struct {
	Message      CopilotMessage `json:"message"`
	LocationHint LocationHint   `json:"locationHint"`
	// Optional fields for future implementation
	// AdditionalContext []CopilotContextMessage `json:"additionalContext,omitempty"`
	// ContextualResources *CopilotContextualResources `json:"contextualResources,omitempty"`
}

// CopilotContextMessage represents additional context for Copilot
type CopilotContextMessage struct {
	Text string `json:"text"`
}

// CopilotContextualResources provides context files and web search settings
type CopilotContextualResources struct {
	Files      []CopilotFile `json:"files,omitempty"`
	WebContext *WebContext   `json:"webContext,omitempty"`
}

// CopilotFile represents a file reference for context
type CopilotFile struct {
	URI string `json:"uri"`
}

// WebContext controls web search grounding
type WebContext struct {
	IsWebEnabled bool `json:"isWebEnabled"`
}

// CopilotConversationResponse represents the response from creating a conversation
type ConversationResponse struct {
	ID              string                       `json:"id"`
	CreatedDateTime time.Time                    `json:"createdDateTime"`
	DisplayName     string                       `json:"displayName"`
	State           string                       `json:"state"`
	TurnCount       int                          `json:"turnCount"`
	Messages        []CopilotConversationMessage `json:"messages,omitempty"`
}

// CopilotConversationMessage represents a message in a conversation
type CopilotConversationMessage struct {
	ID               string                   `json:"id"`
	Text             string                   `json:"text"`
	CreatedDateTime  time.Time                `json:"createdDateTime"`
	AdaptiveCards    []interface{}            `json:"adaptiveCards"`
	Attributions     []CopilotAttribution     `json:"attributions"`
	SensitivityLabel *CopilotSensitivityLabel `json:"sensitivityLabel"`
}

// CopilotAttribution represents source attribution for responses
type CopilotAttribution struct {
	AttributionType     string `json:"attributionType"`
	ProviderDisplayName string `json:"providerDisplayName"`
	AttributionSource   string `json:"attributionSource"`
	SeeMoreWebURL       string `json:"seeMoreWebUrl"`
	ImageWebURL         string `json:"imageWebUrl"`
	ImageFavIcon        string `json:"imageFavIcon"`
	ImageWidth          int    `json:"imageWidth"`
	ImageHeight         int    `json:"imageHeight"`
}

// CopilotSensitivityLabel represents sensitivity information
type CopilotSensitivityLabel struct {
	SensitivityLabelID string `json:"sensitivityLabelId"`
	DisplayName        string `json:"displayName"`
	Tooltip            string `json:"tooltip"`
	Priority           string `json:"priority"`
	Color              string `json:"color"`
	IsEncrypted        bool   `json:"isEncrypted"`
}
