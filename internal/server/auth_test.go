package restapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danielmiessler/fabric/internal/core"
	"github.com/danielmiessler/fabric/internal/plugins/db/fsdb"
	"github.com/gin-gonic/gin"
)

func TestAPIKeyMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(APIKeyMiddleware("secret"))
	router.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	router.GET("/swagger/*any", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	tests := []struct {
		name       string
		path       string
		header     string
		wantStatus int
	}{
		{
			name:       "protected route rejects missing api key",
			path:       "/protected",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "protected route accepts correct api key",
			path:       "/protected",
			header:     "secret",
			wantStatus: http.StatusOK,
		},
		{
			name:       "protected route rejects wrong api key",
			path:       "/protected",
			header:     "secret-extra",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "swagger route remains public",
			path:       "/swagger/index.html",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			if tt.header != "" {
				req.Header.Set(APIKeyHeader, tt.header)
			}

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, recorder.Code)
			}
		})
	}
}

func TestRESTServerRoutesRequireAPIKeyWhenConfigured(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := fsdb.NewDb(t.TempDir())
	registry := &core.PluginRegistry{Db: db}

	router := gin.New()
	router.Use(APIKeyMiddleware("secret"))
	NewPatternsHandler(router, db.Patterns)
	NewContextsHandler(router, db.Contexts)
	NewSessionsHandler(router, db.Sessions)
	NewChatHandler(router, registry, db)
	NewYouTubeHandler(router, registry)
	NewConfigHandler(router, db)
	NewModelsHandler(router, nil)
	NewStrategiesHandler(router)

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{name: "chat", method: http.MethodPost, path: "/chat"},
		{name: "patterns", method: http.MethodGet, path: "/patterns/names"},
		{name: "contexts", method: http.MethodGet, path: "/contexts/names"},
		{name: "sessions", method: http.MethodGet, path: "/sessions/names"},
		{name: "youtube", method: http.MethodPost, path: "/youtube/transcript"},
		{name: "config", method: http.MethodGet, path: "/config"},
		{name: "models", method: http.MethodGet, path: "/models/names"},
		{name: "strategies", method: http.MethodGet, path: "/strategies"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			if recorder.Code != http.StatusUnauthorized {
				t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, recorder.Code)
			}
		})
	}
}
