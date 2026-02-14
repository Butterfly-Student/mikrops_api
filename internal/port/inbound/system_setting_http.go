package inbound_port

import "github.com/gin-gonic/gin"

type SystemSettingHttpPort interface {
	GetByKey(c *gin.Context)
	GetAll(c *gin.Context)
	UpdateByKey(c *gin.Context)
	GetPublic(c *gin.Context)
}
