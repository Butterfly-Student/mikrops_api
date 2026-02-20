package inbound_port

import "github.com/gin-gonic/gin"

type PaymentHttpPort interface {
	Create(c *gin.Context)
	GetByID(c *gin.Context)
	GetByNumber(c *gin.Context)
	List(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	Confirm(c *gin.Context)
	Reject(c *gin.Context)
	Allocate(c *gin.Context)
}
