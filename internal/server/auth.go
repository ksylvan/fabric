package restapi

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const APIKeyHeader = "X-API-Key"

// APIKeyMiddleware validates API key for protected endpoints.
// Swagger documentation endpoints (/swagger/*) are exempt from authentication
// to allow users to browse and test the API documentation freely.
func APIKeyMiddleware(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip authentication for Swagger documentation endpoints
		// This allows public access to API docs even when authentication is enabled
		if strings.HasPrefix(c.Request.URL.Path, "/swagger/") {
			c.Next()
			return
		}

		headerApiKey := c.GetHeader(APIKeyHeader)

		if headerApiKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing API Key"})
			return
		}

		if subtle.ConstantTimeCompare([]byte(headerApiKey), []byte(apiKey)) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Wrong API Key"})
			return
		}

		c.Next()
	}
}
