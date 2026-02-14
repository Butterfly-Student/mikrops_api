package inbound_port

import "github.com/gin-gonic/gin"

type IpPoolHttpPort interface {
	// CRUD
	CreateIpPool(c *gin.Context)
	UpdateIpPool(c *gin.Context)
	DeleteIpPool(c *gin.Context)
	GetIpPool(c *gin.Context)
	ListIpPools(c *gin.Context)
}
