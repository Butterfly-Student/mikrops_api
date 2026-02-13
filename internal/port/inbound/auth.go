package inbound_port

import "github.com/gin-gonic/gin"

type AuthHttpPort interface {
	Login(c *gin.Context)
	Register(c *gin.Context)
	RefreshToken(c *gin.Context)
	ChangePassword(c *gin.Context)
	Logout(c *gin.Context)
}
