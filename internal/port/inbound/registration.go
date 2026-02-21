package inbound_port

import "github.com/gin-gonic/gin"

type RegistrationHttpPort interface {
	Submit(c *gin.Context)
	GetByID(c *gin.Context)
	List(c *gin.Context)
	Approve(c *gin.Context)
	Reject(c *gin.Context)
}
