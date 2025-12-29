package restapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// respondWithError sends a JSON error response with the given HTTP status code
func respondWithError(c *gin.Context, statusCode int, err error) {
	c.JSON(statusCode, err.Error())
}

// respondWithInternalError sends a 500 Internal Server Error JSON response
func respondWithInternalError(c *gin.Context, err error) {
	respondWithError(c, http.StatusInternalServerError, err)
}

// respondWithBadRequest sends a 400 Bad Request JSON response
func respondWithBadRequest(c *gin.Context, err error) {
	respondWithError(c, http.StatusBadRequest, err)
}
