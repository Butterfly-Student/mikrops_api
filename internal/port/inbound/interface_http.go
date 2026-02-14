package inbound_port

import "github.com/gin-gonic/gin"

type InterfaceHttpPort interface {
	// Monitoring
	StartMonitoring(c *gin.Context)
	StartMonitoringByName(c *gin.Context)
	StopMonitoring(c *gin.Context)
	StopMonitoringByName(c *gin.Context)

	// WebSocket
	HandleWebSocket(c *gin.Context)
}
