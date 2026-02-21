package inbound_port

import "github.com/gin-gonic/gin"

type CustomerPortalHttpPort interface {
	Login(c *gin.Context)
	RefreshToken(c *gin.Context)
	GetProfile(c *gin.Context)
	UpdateProfile(c *gin.Context)
	ChangePassword(c *gin.Context)
	ChangePppCredentials(c *gin.Context)
	ListInvoices(c *gin.Context)
	GetInvoice(c *gin.Context)
}
