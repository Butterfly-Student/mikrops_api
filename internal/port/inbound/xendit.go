package inbound_port

import "github.com/gin-gonic/gin"

type XenditHttpPort interface {
	HandleWebhook(c *gin.Context)
}
