package inbound_port

import "github.com/gin-gonic/gin"

type QueueHttpPort interface {
	// CRUD
	CreateQueue(c *gin.Context)
	UpdateQueue(c *gin.Context)
	DeleteQueue(c *gin.Context)
	GetQueue(c *gin.Context)
	ListQueues(c *gin.Context)

	// Streaming
	StartStreaming(c *gin.Context)
	HandleWebSocket(c *gin.Context)
}
