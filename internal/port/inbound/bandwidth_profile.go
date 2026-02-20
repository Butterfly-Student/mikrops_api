package inbound_port

import "github.com/gin-gonic/gin"

type BandwidthProfileHttpPort interface {
	Create(c *gin.Context)
	CreateWithRouter(c *gin.Context) // MikroTik-first: POST /mikrotik/:router_id/bandwidth-profiles
	GetByID(c *gin.Context)
	GetByCode(c *gin.Context)
	List(c *gin.Context)
	Update(c *gin.Context)
	UpdateWithRouter(c *gin.Context) // MikroTik-first: PUT /mikrotik/:router_id/bandwidth-profiles/:id
	Delete(c *gin.Context)
	DeleteWithRouter(c *gin.Context) // MikroTik-first: DELETE /mikrotik/:router_id/bandwidth-profiles/:id
	SyncToMikrotik(c *gin.Context)
}
