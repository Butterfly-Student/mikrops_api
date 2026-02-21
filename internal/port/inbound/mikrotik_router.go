package inbound_port

import "github.com/gin-gonic/gin"

// MikrotikRouterHttpPort defines HTTP handlers for MikroTik router management
//
//go:generate mockgen -source=mikrotik_router.go -destination=./../../../tests/mocks/port/mock_mikrotik_router_http.go
type MikrotikRouterHttpPort interface {
	Create(c *gin.Context)
	List(c *gin.Context)
	GetByID(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	TestConnection(c *gin.Context)
	SetupIsolation(c *gin.Context)
	CheckIsolationSetup(c *gin.Context)
}
