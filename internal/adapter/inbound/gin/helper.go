package gin_inbound_adapter

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// getRouterID retrieves router_id from query parameter
func getRouterID(c *gin.Context) (string, error) {
	idStr := c.Query("router_id")
	if idStr == "" {
		return "", errors.New("router_id query parameter required")
	}
	_, err := uuid.Parse(idStr)
	if err != nil {
		return "", errors.New("invalid router_id format")
	}
	return idStr, nil
}

// respondWithError sends error response
func respondWithError(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{"error": message})
}

// respondWithSuccess sends success response
func respondWithSuccess(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{"message": message})
}

// respondWithJSON sends JSON response
func respondWithJSON(c *gin.Context, code int, obj any) {
	c.JSON(code, obj)
}
