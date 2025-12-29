package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/danielmiessler/fabric/internal/i18n"
	debuglog "github.com/danielmiessler/fabric/internal/log"
	"github.com/danielmiessler/fabric/internal/plugins/ai/helpers"
)

// modelResponse represents a minimal model returned by the API.
// This mirrors the shape used by OpenAI-compatible providers that return
// either an array of models or an object with a `data` field.
type modelResponse struct {
	ID string `json:"id"`
}

// FetchModelsDirectly is used to fetch models directly from the API when the
// standard OpenAI SDK method fails due to a nonstandard format. This is useful
// for providers that return a direct array of models (e.g., GitHub Models) or
// other OpenAI-compatible implementations.
// If httpClient is nil, a new client with default settings will be created.
func FetchModelsDirectly(ctx context.Context, baseURL, apiKey, providerName string, httpClient *http.Client) ([]string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if baseURL == "" {
		return nil, fmt.Errorf(i18n.T("openai_api_base_url_not_configured"), providerName)
	}

	// Build the /models endpoint URL
	fullURL, err := url.JoinPath(baseURL, "models")
	if err != nil {
		return nil, fmt.Errorf(i18n.T("openai_failed_to_create_models_url"), err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	req.Header.Set("Accept", "application/json")

	// Reuse provided HTTP client, or create a new one if not provided
	client := httpClient
	if client == nil {
		client = &http.Client{
			Timeout: defaultHTTPTimeout,
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check HTTP status and get error details if not successful
	if err := helpers.CheckHTTPStatus(resp, helpers.DefaultErrorBodyLimit); err != nil {
		return nil, fmt.Errorf("%s: %w", providerName, err)
	}

	// Read the response body with size validation to prevent memory exhaustion
	bodyBytes, err := helpers.ValidateResponseSize(resp.Body, helpers.DefaultMaxResponseSize)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", providerName, err)
	}

	// Try to parse as an object with data field (OpenAI format)
	var openAIFormat struct {
		Data []modelResponse `json:"data"`
	}
	// Try to parse as a direct array
	var directArray []modelResponse

	if err := json.Unmarshal(bodyBytes, &openAIFormat); err == nil {
		debuglog.Debug(debuglog.Detailed, "Successfully parsed models response from %s using OpenAI format (found %d models)\n", providerName, len(openAIFormat.Data))
		return extractModelIDs(openAIFormat.Data), nil
	}

	if err := json.Unmarshal(bodyBytes, &directArray); err == nil {
		debuglog.Debug(debuglog.Detailed, "Successfully parsed models response from %s using direct array format (found %d models)\n", providerName, len(directArray))
		return extractModelIDs(directArray), nil
	}

	// Truncate error body for readability
	var truncatedBody string
	if len(bodyBytes) > helpers.DefaultErrorBodyLimit {
		truncatedBody = string(bodyBytes[:helpers.DefaultErrorBodyLimit]) + "..."
	} else {
		truncatedBody = string(bodyBytes)
	}
	return nil, fmt.Errorf(i18n.T("openai_unable_to_parse_models_response"), truncatedBody)
}

func extractModelIDs(models []modelResponse) []string {
	modelIDs := make([]string, 0, len(models))
	for _, model := range models {
		modelIDs = append(modelIDs, model.ID)
	}
	return modelIDs
}
