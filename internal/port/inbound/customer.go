package inbound_port

import "github.com/gin-gonic/gin"

type CustomerHttpPort interface {
	// Read-only (no MikroTik required)
	GetByID(c *gin.Context)
	GetByCode(c *gin.Context)
	List(c *gin.Context)

	// MikroTik-first (under /mikrotik/:router_id/customers)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	ChangeStatus(c *gin.Context)
	Isolate(c *gin.Context)
	UnIsolate(c *gin.Context)
	SyncToMikrotik(c *gin.Context)
}
