package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-template/internal/domain"
	"go-template/internal/model"
)

func ActivityLoggingMiddleware(domain domain.Domain) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)

		activityLog := &model.ActivityLog{
			Action:      c.Request.Method + " " + c.Request.URL.Path,
			EntityType:  "api_request",
			Description: fmt.Sprintf("API call completed in %v", duration),
			IPAddress:   c.ClientIP(),
			UserAgent:   c.Request.UserAgent(),
		}

		userID := getUserIDFromContext(c)
		if userID != nil {
			activityLog.UserID = userID
		}

		// Log asynchronously
		go domain.Activity().LogActivity(c.Request.Context(), model.ActivityLogInput{
			UserID:      activityLog.UserID,
			Action:      activityLog.Action,
			EntityType:  activityLog.EntityType,
			Description: activityLog.Description,
			IPAddress:   activityLog.IPAddress,
			UserAgent:   activityLog.UserAgent,
		})
	}
}

func getUserIDFromContext(c *gin.Context) *uuid.UUID {
	userID, exists := c.Get("user_id")
	if !exists {
		return nil
	}

	if userIDStr, ok := userID.(string); ok {
		parsedID, err := uuid.Parse(userIDStr)
		if err == nil {
			return &parsedID
		}
	}

	return nil
}