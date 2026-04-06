package restapi

import (
	"net/http"
	"strings"

	"github.com/danielmiessler/fabric/internal/plugins/db/fsdb"
	"github.com/gin-gonic/gin"
)

// ConfigHandler defines the handler for configuration-related operations
type ConfigHandler struct {
	db *fsdb.Db
	// configurations *fsdb.EnvFilePath("$HOME/.config/fabric/.env")
}

type configUpdateRequest struct {
	OpenAIApiKey     *string `json:"openai_api_key"`
	AnthropicApiKey  *string `json:"anthropic_api_key"`
	GroqApiKey       *string `json:"groq_api_key"`
	MistralApiKey    *string `json:"mistral_api_key"`
	GeminiApiKey     *string `json:"gemini_api_key"`
	OllamaURL        *string `json:"ollama_url"`
	OpenRouterApiKey *string `json:"openrouter_api_key"`
	SiliconApiKey    *string `json:"silicon_api_key"`
	DeepSeekApiKey   *string `json:"deepseek_api_key"`
	GrokaiApiKey     *string `json:"grokai_api_key"`
	LMStudioURL      *string `json:"lm_studio_base_url"`
}

func NewConfigHandler(r *gin.Engine, db *fsdb.Db) *ConfigHandler {
	handler := &ConfigHandler{
		db: db,
		// configurations: db.Configurations,
	}

	r.GET("/config", handler.GetConfig)
	r.POST("/config/update", handler.UpdateConfig)

	return handler
}

// maskAPIKey redacts all but the last 4 characters of a secret key (CWE-200).
// An empty value (key not configured) is returned unchanged so the UI can
// distinguish "not set" from "set but redacted".
func maskAPIKey(key string) string {
	const visible = 4
	if len(key) <= visible {
		return key
	}
	return strings.Repeat("*", len(key)-visible) + key[len(key)-visible:]
}

// isRedacted returns true when a submitted value looks like a masked key
// returned by maskAPIKey, signalling that the user did not change the field.
func isRedacted(value string) bool {
	return strings.Contains(value, "*")
}

func (h *ConfigHandler) GetConfig(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": ".env file not found"})
		return
	}

	envVars, err := h.db.LoadEnvMap()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// API keys are masked to their last 4 characters (CWE-200).
	// URLs are not secrets and are returned as-is so the UI can display them.
	config := map[string]string{
		"openai":     maskAPIKey(envVars["OPENAI_API_KEY"]),
		"anthropic":  maskAPIKey(envVars["ANTHROPIC_API_KEY"]),
		"groq":       maskAPIKey(envVars["GROQ_API_KEY"]),
		"mistral":    maskAPIKey(envVars["MISTRAL_API_KEY"]),
		"gemini":     maskAPIKey(envVars["GEMINI_API_KEY"]),
		"ollama":     envVars["OLLAMA_URL"],
		"openrouter": maskAPIKey(envVars["OPENROUTER_API_KEY"]),
		"silicon":    maskAPIKey(envVars["SILICON_API_KEY"]),
		"deepseek":   maskAPIKey(envVars["DEEPSEEK_API_KEY"]),
		"grokai":     maskAPIKey(envVars["GROKAI_API_KEY"]),
		"lmstudio":   envVars["LM_STUDIO_API_BASE_URL"],
	}

	c.JSON(http.StatusOK, config)
}

func (h *ConfigHandler) UpdateConfig(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database not initialized"})
		return
	}

	var config configUpdateRequest
	if err := decodeStrictJSON(c, &config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]*string{
		"OPENAI_API_KEY":         requestValueUpdate(config.OpenAIApiKey),
		"ANTHROPIC_API_KEY":      requestValueUpdate(config.AnthropicApiKey),
		"GROQ_API_KEY":           requestValueUpdate(config.GroqApiKey),
		"MISTRAL_API_KEY":        requestValueUpdate(config.MistralApiKey),
		"GEMINI_API_KEY":         requestValueUpdate(config.GeminiApiKey),
		"OLLAMA_URL":             requestValueUpdate(config.OllamaURL),
		"OPENROUTER_API_KEY":     requestValueUpdate(config.OpenRouterApiKey),
		"SILICON_API_KEY":        requestValueUpdate(config.SiliconApiKey),
		"DEEPSEEK_API_KEY":       requestValueUpdate(config.DeepSeekApiKey),
		"GROKAI_API_KEY":         requestValueUpdate(config.GrokaiApiKey),
		"LM_STUDIO_API_BASE_URL": requestValueUpdate(config.LMStudioURL),
	}

	if err := h.db.UpdateEnvValues(updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Configuration updated successfully"})
}

func requestValueUpdate(value *string) *string {
	if value == nil {
		return nil
	}
	if isRedacted(*value) {
		return nil
	}
	return value
}
