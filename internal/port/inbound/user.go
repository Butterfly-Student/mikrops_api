package inbound_port

import "github.com/gin-gonic/gin"

type UserHttpPort interface {
	GetProfile(c *gin.Context)
	UpdateProfile(c *gin.Context)
}
