package inbound_port

import "github.com/gin-gonic/gin"

type BandwidthProfileHttpPort interface {
	CreateProfile(c *gin.Context)
	GetProfile(c *gin.Context)
	ListProfiles(c *gin.Context)
	UpdateProfile(c *gin.Context)
	DeleteProfile(c *gin.Context)
	SyncToMikrotik(c *gin.Context)
	GetIsolatedProfile(c *gin.Context)
}
