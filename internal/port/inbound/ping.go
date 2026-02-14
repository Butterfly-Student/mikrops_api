package inbound_port

import "github.com/gin-gonic/gin"

type PingHttpPort interface {
	GetResource(c *gin.Context)

	// Ping Operations
	StartPing(c *gin.Context)
	StopPing(c *gin.Context)

	// WebSocket
	HandleWebSocket(c *gin.Context)
}
