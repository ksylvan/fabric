package azure

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/danielmiessler/fabric/internal/i18n"
	"github.com/danielmiessler/fabric/internal/plugins"
	"github.com/danielmiessler/fabric/internal/plugins/ai/openai"
	openaiapi "github.com/openai/openai-go"
	"github.com/openai/openai-go/azure"
	"github.com/openai/openai-go/option"
)

func NewClient() (ret *Client) {
	ret = &Client{}
	ret.Client = openai.NewClientCompatible("Azure", "", ret.configure)
	ret.ApiDeployments = ret.AddSetupQuestionCustom("deployments", true,
		i18n.T("azure_deployments_question"))
	ret.ApiVersion = ret.AddSetupQuestionCustom("API Version", false,
		i18n.T("azure_api_version_question"))

	return
}

type Client struct {
	*openai.Client
	ApiDeployments *plugins.SetupQuestion
	ApiVersion     *plugins.SetupQuestion

	apiDeployments []string
}

const defaultAPIVersion = "2025-04-01-preview"

func (oi *Client) configure() error {
	oi.apiDeployments = parseDeployments(oi.ApiDeployments.Value)

	apiKey := strings.TrimSpace(oi.ApiKey.Value)
	if apiKey == "" {
		return fmt.Errorf("%s", i18n.T("azure_api_key_required"))
	}

	baseURL := strings.TrimSpace(oi.ApiBaseURL.Value)
	if baseURL == "" {
		return fmt.Errorf("%s", i18n.T("azure_base_url_required"))
	}

	apiVersion := strings.TrimSpace(oi.ApiVersion.Value)
	if apiVersion == "" {
		apiVersion = defaultAPIVersion
		oi.ApiVersion.Value = apiVersion
	}

	// Build the Azure endpoint URL with /openai/ suffix
	endpoint := strings.TrimSuffix(baseURL, "/") + "/openai/"

	// Create the client with Azure authentication and custom middleware
	// to fix the deployment URL path (workaround for SDK bug where
	// jsonRoutes expects /openai/chat/completions but SDK uses /chat/completions)
	client := openaiapi.NewClient(
		azure.WithAPIKey(apiKey),
		option.WithBaseURL(endpoint),
		option.WithQueryAdd("api-version", apiVersion),
		option.WithMiddleware(azureDeploymentMiddleware),
	)
	oi.ApiClient = &client
	return nil
}

// azureDeploymentMiddleware transforms Azure OpenAI API paths to include
// the deployment name. Azure requires URLs like:
// /openai/deployments/{deployment-name}/chat/completions
// but the SDK sends paths like /chat/completions
func azureDeploymentMiddleware(req *http.Request, next option.MiddlewareNext) (*http.Response, error) {
	// Routes that need deployment name injection
	deploymentRoutes := map[string]bool{
		"/chat/completions":     true,
		"/completions":          true,
		"/embeddings":           true,
		"/audio/speech":         true,
		"/audio/transcriptions": true,
		"/audio/translations":   true,
		"/images/generations":   true,
		"/responses":            true,
	}

	path := req.URL.Path

	// Remove /openai prefix if present (SDK may add it via base URL)
	trimmedPath := strings.TrimPrefix(path, "/openai")
	if !strings.HasPrefix(trimmedPath, "/") {
		trimmedPath = "/" + trimmedPath
	}

	if deploymentRoutes[trimmedPath] {
		// Extract model/deployment name from request body
		deploymentName, err := extractDeploymentFromBody(req)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", i18n.T("azure_failed_extract_deployment"), err)
		}

		// Transform path: /chat/completions -> /deployments/{name}/chat/completions
		newPath := "/openai/deployments/" + url.PathEscape(deploymentName) + trimmedPath
		req.URL.Path = newPath
		req.URL.RawPath = "" // Clear RawPath to ensure Path is used
	}

	return next(req)
}

// extractDeploymentFromBody reads the model field from the JSON request body
// and restores the body for subsequent use
func extractDeploymentFromBody(req *http.Request) (string, error) {
	if req.Body == nil {
		return "", fmt.Errorf("%s", i18n.T("azure_request_body_nil"))
	}

	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return "", err
	}
	// Restore body for subsequent reads
	req.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	var payload struct {
		Model string `json:"model"`
	}
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		return "", err
	}

	if payload.Model == "" {
		return "", fmt.Errorf("%s", i18n.T("azure_model_field_empty"))
	}

	return payload.Model, nil
}

func parseDeployments(value string) []string {
	parts := strings.Split(value, ",")
	var deployments []string
	for _, part := range parts {
		if deployment := strings.TrimSpace(part); deployment != "" {
			deployments = append(deployments, deployment)
		}
	}
	return deployments
}

func (oi *Client) ListModels() (ret []string, err error) {
	ret = oi.apiDeployments
	return
}
