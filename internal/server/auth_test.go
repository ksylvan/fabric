package restapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

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
