package inbound_port

import "github.com/gin-gonic/gin"

type MikrotikHttpPort interface {
	CreateRouter(c *gin.Context)
	FindRouterByID(c *gin.Context)
	ListRouters(c *gin.Context)
	UpdateRouter(c *gin.Context)
	DeleteRouter(c *gin.Context)

	SetActiveRouter(c *gin.Context)
	GetActiveRouter(c *gin.Context)

	TestRouterConnection(c *gin.Context)
}
