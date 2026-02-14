package inbound_port

import "github.com/gin-gonic/gin"

type CustomerHttpPort interface {
	Create(c *gin.Context)
	Get(c *gin.Context)
	List(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	GetBillingInfo(c *gin.Context)
	Isolate(c *gin.Context)
	Activate(c *gin.Context)
}
