package inbound_port

import "github.com/gin-gonic/gin"

type PppoeHttpPort interface {
	// Secret Management
	CreateSecret(c *gin.Context)
	UpdateSecret(c *gin.Context)
	DeleteSecret(c *gin.Context)
	GetSecret(c *gin.Context)
	ListSecrets(c *gin.Context)

	// Profile Management
	CreateProfile(c *gin.Context)
	UpdateProfile(c *gin.Context)
	DeleteProfile(c *gin.Context)
	GetProfile(c *gin.Context)
	ListProfiles(c *gin.Context)

	// Session Management (Active/Inactive)
	ListActiveSessions(c *gin.Context)
	ListInactiveSessions(c *gin.Context)
	ListSessionHistory(c *gin.Context)

	// Webhooks
	CallbackOnUp(c *gin.Context)
	CallbackOnDown(c *gin.Context)

	// WebSocket
	HandleWebSocket(c *gin.Context)
}
