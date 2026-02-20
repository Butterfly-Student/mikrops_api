package inbound_port

import "github.com/gin-gonic/gin"

type CustomerHttpPort interface {
	Create(c *gin.Context)
	GetByID(c *gin.Context)
	GetByCode(c *gin.Context)
	List(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	ChangeStatus(c *gin.Context)
	Isolate(c *gin.Context)
	UnIsolate(c *gin.Context)
	SyncToMikrotik(c *gin.Context)
}
