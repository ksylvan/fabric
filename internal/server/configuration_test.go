package restapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/danielmiessler/fabric/internal/plugins/db/fsdb"
	"github.com/gin-gonic/gin"
)

func TestConfigHandlerGetConfigUsesEnvFileAndMasksSecrets(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Setenv("OPENAI_API_KEY", "sk-process-9999")

	db := fsdb.NewDb(t.TempDir())
	if err := db.SaveEnvMap(map[string]string{
		"OPENAI_API_KEY": "sk-file-1234",
		"OLLAMA_URL":     "http://localhost:11434",
		"DEFAULT_VENDOR": "openai",
	}); err != nil {
		t.Fatalf("SaveEnvMap() error = %v", err)
	}

	router := gin.New()
	NewConfigHandler(router, db)

	req := httptest.NewRequest(http.MethodGet, "/config", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var got map[string]string
	if err := json.NewDecoder(recorder.Body).Decode(&got); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	if got["openai"] != maskAPIKey("sk-file-1234") {
		t.Fatalf("openai mask = %q, want %q", got["openai"], maskAPIKey("sk-file-1234"))
	}
	if got["openai"] == maskAPIKey("sk-process-9999") {
		t.Fatalf("openai mask unexpectedly came from process env, got %q", got["openai"])
	}
	if got["ollama"] != "http://localhost:11434" {
		t.Fatalf("ollama = %q, want %q", got["ollama"], "http://localhost:11434")
	}
	if got["anthropic"] != "" {
		t.Fatalf("anthropic = %q, want empty", got["anthropic"])
	}
	if _, ok := got["DEFAULT_VENDOR"]; ok {
		t.Fatalf("response leaked unrelated env entry: %#v", got)
	}
}

func TestConfigHandlerUpdateConfigPreservesMaskedSecretsAndAllowsClear(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := fsdb.NewDb(t.TempDir())
	initial := map[string]string{
		"OPENAI_API_KEY":    "sk-openai-1234",
		"ANTHROPIC_API_KEY": "sk-ant-old",
		"DEFAULT_VENDOR":    "Anthropic",
		"OLLAMA_URL":        "http://localhost:11434",
	}
	if err := db.SaveEnvMap(initial); err != nil {
		t.Fatalf("SaveEnvMap() error = %v", err)
	}

	t.Setenv("OPENAI_API_KEY", initial["OPENAI_API_KEY"])
	t.Setenv("ANTHROPIC_API_KEY", initial["ANTHROPIC_API_KEY"])
	t.Setenv("OLLAMA_URL", initial["OLLAMA_URL"])

	router := gin.New()
	NewConfigHandler(router, db)

	body := `{"openai_api_key":"` + maskAPIKey(initial["OPENAI_API_KEY"]) + `","anthropic_api_key":"sk-ant-new","ollama_url":""}`
	req := httptest.NewRequest(http.MethodPost, "/config/update", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	envMap, err := db.LoadEnvMap()
	if err != nil {
		t.Fatalf("LoadEnvMap() error = %v", err)
	}

	if envMap["OPENAI_API_KEY"] != initial["OPENAI_API_KEY"] {
		t.Fatalf("OPENAI_API_KEY = %q, want preserved %q", envMap["OPENAI_API_KEY"], initial["OPENAI_API_KEY"])
	}
	if envMap["ANTHROPIC_API_KEY"] != "sk-ant-new" {
		t.Fatalf("ANTHROPIC_API_KEY = %q, want %q", envMap["ANTHROPIC_API_KEY"], "sk-ant-new")
	}
	if _, ok := envMap["OLLAMA_URL"]; ok {
		t.Fatalf("OLLAMA_URL should be cleared, got %q", envMap["OLLAMA_URL"])
	}
	if envMap["DEFAULT_VENDOR"] != "Anthropic" {
		t.Fatalf("DEFAULT_VENDOR = %q, want %q", envMap["DEFAULT_VENDOR"], "Anthropic")
	}

	if got := os.Getenv("ANTHROPIC_API_KEY"); got != "sk-ant-new" {
		t.Fatalf("process env ANTHROPIC_API_KEY = %q, want %q", got, "sk-ant-new")
	}
	if got := os.Getenv("OLLAMA_URL"); got != "" {
		t.Fatalf("process env OLLAMA_URL = %q, want cleared", got)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/config", nil)
	getRecorder := httptest.NewRecorder()
	router.ServeHTTP(getRecorder, getReq)

	if getRecorder.Code != http.StatusOK {
		t.Fatalf("GET /config status = %d, want %d", getRecorder.Code, http.StatusOK)
	}

	var got map[string]string
	if err := json.NewDecoder(getRecorder.Body).Decode(&got); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if got["openai"] != maskAPIKey(initial["OPENAI_API_KEY"]) {
		t.Fatalf("GET /config openai = %q, want %q", got["openai"], maskAPIKey(initial["OPENAI_API_KEY"]))
	}
	if got["anthropic"] != maskAPIKey("sk-ant-new") {
		t.Fatalf("GET /config anthropic = %q, want %q", got["anthropic"], maskAPIKey("sk-ant-new"))
	}
	if got["ollama"] != "" {
		t.Fatalf("GET /config ollama = %q, want cleared", got["ollama"])
	}
}
