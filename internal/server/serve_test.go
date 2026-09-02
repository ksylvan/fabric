package restapi

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/danielmiessler/fabric/internal/core"
	"github.com/danielmiessler/fabric/internal/plugins/db/fsdb"
	"github.com/gin-gonic/gin"
)

func TestNewServeEngine_APIKeyWiring(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := fsdb.NewDb(t.TempDir())
	if err := os.MkdirAll(db.Contexts.Dir, 0o755); err != nil {
		t.Fatalf("MkdirAll(contexts) error = %v", err)
	}

	registry := &core.PluginRegistry{Db: db}
	engine := newServeEngine(registry, "secret")

	request := func(method, path, key, body string) *httptest.ResponseRecorder {
		var reader *strings.Reader
		if body == "" {
			reader = strings.NewReader("")
		} else {
			reader = strings.NewReader(body)
		}
		req := httptest.NewRequest(method, path, reader)
		if body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		if key != "" {
			req.Header.Set(APIKeyHeader, key)
		}
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)
		return w
	}

	if got := request(http.MethodGet, "/contexts/names", "", "").Code; got != http.StatusUnauthorized {
		t.Fatalf("GET /contexts/names without key = %d, want 401", got)
	}
	if got := request(http.MethodPost, "/config/update", "", `{"openai_api_key":"test-key"}`).Code; got != http.StatusUnauthorized {
		t.Fatalf("POST /config/update without key = %d, want 401", got)
	}

	swagger := request(http.MethodGet, "/swagger/index.html", "", "")
	if swagger.Code == http.StatusUnauthorized {
		t.Fatalf("GET /swagger/index.html without key = %d, want non-401 public docs route", swagger.Code)
	}

	if got := request(http.MethodGet, "/contexts/names", "secret", "").Code; got != http.StatusOK {
		t.Fatalf("GET /contexts/names with valid key = %d, want 200", got)
	}
	if got := request(http.MethodPost, "/config/update", "secret", `{"openai_api_key":"test-key"}`).Code; got != http.StatusOK {
		t.Fatalf("POST /config/update with valid key = %d, want 200", got)
	}
}
