package inbound_port

import "github.com/gin-gonic/gin"

type SystemSettingHttpPort interface {
	GetSetting(c *gin.Context)
	ListSettings(c *gin.Context)
	UpdateSetting(c *gin.Context)
}
