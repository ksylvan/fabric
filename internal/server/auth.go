package restapi

import (
	"crypto/sha256"
	"crypto/subtle"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const APIKeyHeader = "X-API-Key"

func logSecurityEvent(c *gin.Context, message string, attrs ...any) {
	baseAttrs := []any{
		"method", c.Request.Method,
		"path", c.Request.URL.Path,
		"client_ip", c.ClientIP(),
	}
	if route := c.FullPath(); route != "" {
		baseAttrs = append(baseAttrs, "route", route)
	}
	slog.Warn(message, append(baseAttrs, attrs...)...)
}

// APIKeyMiddleware validates API key for protected endpoints.
// Swagger documentation endpoints (/swagger/*) are exempt from authentication
// to allow users to browse and test the API documentation freely.
func APIKeyMiddleware(apiKey string) gin.HandlerFunc {
	// Compare digests, not the raw values: ConstantTimeCompare returns
	// early on a length mismatch, which leaks the configured key length.
	expectedKey := sha256.Sum256([]byte(apiKey))
	return func(c *gin.Context) {
		// Skip authentication for Swagger documentation endpoints
		// This allows public access to API docs even when authentication is enabled
		if strings.HasPrefix(c.Request.URL.Path, "/swagger/") {
			c.Next()
			return
		}

		headerApiKey := c.GetHeader(APIKeyHeader)

		if headerApiKey == "" {
			logSecurityEvent(c, "API key authentication failed", "reason", "missing_api_key")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing API Key"})
			return
		}

		headerKey := sha256.Sum256([]byte(headerApiKey))
		if subtle.ConstantTimeCompare(headerKey[:], expectedKey[:]) != 1 {
			logSecurityEvent(c, "API key authentication failed", "reason", "invalid_api_key")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Wrong API Key"})
			return
		}

		c.Next()
	}
}
