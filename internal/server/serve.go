package restapi

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/danielmiessler/fabric/internal/core"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/danielmiessler/fabric/docs" // swagger docs
)

const (
	// swaggerHost is the default host for Swagger documentation
	swaggerHost = "localhost:8080"
)

// @title Fabric REST API
// @version 1.0
// @description REST API for Fabric AI augmentation framework. Provides endpoints for chat completions, pattern management, contexts, sessions, and more.
// @contact.name Fabric Support
// @contact.url https://github.com/danielmiessler/fabric
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
func Serve(registry *core.PluginRegistry, address string, apiKey string) (err error) {
	router := gin.New()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	if apiKey != "" {
		router.Use(APIKeyMiddleware(apiKey))
	} else {
		slog.Warn("Starting REST API server without API key authentication. This may pose security risks.")
	}

	// Swagger UI and documentation endpoint with custom YAML handler
	router.GET("/swagger/*any", func(c *gin.Context) {
		// Check if request is for swagger.yaml
		if c.Param("any") == "/swagger.yaml" {
			// Try to find swagger.yaml relative to current directory or executable
			yamlPath := "docs/swagger.yaml"
			if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
				// Try relative to executable
				if exePath, err := os.Executable(); err == nil {
					yamlPath = filepath.Join(filepath.Dir(exePath), "docs", "swagger.yaml")
				}
			}

			if _, err := os.Stat(yamlPath); err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "swagger.yaml not found - generate it with: swag init -g internal/server/serve.go -o docs"})
				return
			}

			c.File(yamlPath)
			return
		}

		// For all other swagger paths, use the default handler
		ginSwagger.WrapHandler(swaggerFiles.Handler)(c)
	})

	// Register routes
	fabricDb := registry.Db
	NewPatternsHandler(router, fabricDb.Patterns)
	NewContextsHandler(router, fabricDb.Contexts)
	NewSessionsHandler(router, fabricDb.Sessions)
	NewChatHandler(router, registry, fabricDb)
	NewYouTubeHandler(router, registry)
	NewConfigHandler(router, fabricDb)
	NewModelsHandler(router, registry.VendorManager)
	NewStrategiesHandler(router)

	// Start server
	err = router.Run(address)
	if err != nil {
		return err
	}

	return
}
