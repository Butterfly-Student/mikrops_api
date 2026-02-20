package inbound_port

import "github.com/gin-gonic/gin"

type HotspotHttpPort interface {
	// Profile Management
	CreateProfile(c *gin.Context)
	GetProfile(c *gin.Context)
	ListProfiles(c *gin.Context)
	UpdateProfile(c *gin.Context)
	DeleteProfile(c *gin.Context)

	// User Management
	CreateUser(c *gin.Context)
	GetUser(c *gin.Context)
	ListUsers(c *gin.Context)
	UpdateUser(c *gin.Context)
	DeleteUser(c *gin.Context)

	// Voucher Generation
	// GenerateVouchers handles both modes via the "mode" field in the request body:
	// - "vc" (default): username == password  (voucher mode)
	// - "up":           username != password  (user-password mode)
	// Set quantity=1 for a single voucher, quantity=N for bulk.
	GenerateVouchers(c *gin.Context)

	// Session Management
	GetActiveSessions(c *gin.Context)
	GetSessionStats(c *gin.Context)
	DisconnectUser(c *gin.Context)

	// Sales
	RecordSale(c *gin.Context)
	GetSales(c *gin.Context)
	GetTotalRevenue(c *gin.Context)

	// Expiry Schedulers
	CreateExpiryScheduler(c *gin.Context)
	RemoveExpiryScheduler(c *gin.Context)
}
