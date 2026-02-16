package inbound_port

import "github.com/gin-gonic/gin"

type ActivityHttpPort interface {
	ListLogs(c *gin.Context)
	GetEntityHistory(c *gin.Context)
	GetUserLogs(c *gin.Context)
}
