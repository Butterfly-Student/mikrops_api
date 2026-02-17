package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"go-template/internal/domain"
	"go-template/internal/model"
)

func ActivityLoggingMiddleware(domain domain.Domain) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)

		ipAddress := c.ClientIP()
		userAgent := c.Request.UserAgent()
		description := fmt.Sprintf("API call completed in %v", duration)

		activityLog := &model.ActivityLog{
			Action:      c.Request.Method + " " + c.Request.URL.Path,
			EntityType:  "api_request",
			Description: description,
			IPAddress:   &ipAddress,
			UserAgent:   &userAgent,
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

func getUserIDFromContext(c *gin.Context) *uint {
	userID, exists := c.Get("userID")
	if !exists {
		return nil
	}

	if uid, ok := userID.(uint); ok {
		return &uid
	}

	return nil
}
