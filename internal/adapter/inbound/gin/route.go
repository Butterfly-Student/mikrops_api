package gin_inbound_adapter

import (
	"context"

	"github.com/gin-gonic/gin"

	inbound_port "go-template/internal/port/inbound"
)

func InitRoute(
	ctx context.Context,
	app *gin.Engine,
	port inbound_port.HttpPort,
) {
	// Internal routes with internal auth middleware
	internal := app.Group("/internal")
	internal.Use(port.Middleware().InternalAuth())
	{
		internal.POST("/client-upsert", port.Client().Upsert)
		internal.POST("/client-find", port.Client().Find)
		internal.DELETE("/client-delete", port.Client().Delete)
	}

	// V1 routes with client auth middleware
	v1 := app.Group("/v1")
	v1.Use(port.Middleware().ClientAuth())
	{
		v1.GET("/ping", port.Ping().GetResource)
	}

	// Ping Management
	pingMgmt := app.Group("/ping")
	pingMgmt.Use(port.Middleware().UserAuth())
	{
		pingMgmt.POST("", port.Ping().StartPing)
		pingMgmt.DELETE("/:address", port.Ping().StopPing)
	}

	// Ping WebSocket (requires ?address=target filter)
	app.GET("/ws/ping", port.Ping().HandleWebSocket)

	// Auth routes
	auth := app.Group("/auth")
	{
		auth.POST("/login", port.Auth().Login)
		auth.POST("/register", port.Auth().Register)
		auth.POST("/refresh", port.Auth().RefreshToken)
	}

	// User routes
	user := app.Group("/user")
	user.Use(port.Middleware().UserAuth())
	{
		user.POST("/change-password", port.Auth().ChangePassword)
		user.POST("/logout", port.Auth().Logout)
	}

	// Protected user profile routes with RBAC
	userProfile := user.Group("/")
	userProfile.Use(port.Middleware().RBAC())
	{
		userProfile.GET("/profile", port.User().GetProfile)
		userProfile.PUT("/profile", port.User().UpdateProfile)
	}

	// PPPoE Management
	pppoe := app.Group("/pppoe")
	pppoe.Use(port.Middleware().UserAuth())
	// pppoe.Use(port.Middleware().RBAC()) // Enabled RBAC later
	{
		// Secrets
		pppoe.POST("/secrets", port.Pppoe().CreateSecret)
		pppoe.GET("/secrets", port.Pppoe().ListSecrets)
		pppoe.GET("/secrets/:id", port.Pppoe().GetSecret)
		pppoe.PUT("/secrets/:id", port.Pppoe().UpdateSecret)
		pppoe.DELETE("/secrets/:id", port.Pppoe().DeleteSecret)

		// Profiles
		pppoe.POST("/profiles", port.Pppoe().CreateProfile)
		pppoe.GET("/profiles", port.Pppoe().ListProfiles)
		pppoe.GET("/profiles/:id", port.Pppoe().GetProfile)
		pppoe.PUT("/profiles/:id", port.Pppoe().UpdateProfile)
		pppoe.DELETE("/profiles/:id", port.Pppoe().DeleteProfile)

		// Sessions
		pppoe.GET("/sessions/active", port.Pppoe().ListActiveSessions)
		pppoe.GET("/sessions/inactive", port.Pppoe().ListInactiveSessions)
		pppoe.GET("/sessions/history", port.Pppoe().ListSessionHistory)
	}

	// PPPoE Webhooks
	webhooks := app.Group("/webhooks/pppoe")
	{
		webhooks.POST("/on-up", port.Pppoe().CallbackOnUp)
		webhooks.POST("/on-down", port.Pppoe().CallbackOnDown)
	}

	// WebSocket
	app.GET("/ws/pppoe", port.Pppoe().HandleWebSocket)

	// Queue Management
	queue := app.Group("/queues")
	queue.Use(port.Middleware().UserAuth())
	{
		queue.POST("", port.Queue().CreateQueue)
		queue.GET("", port.Queue().ListQueues)
		queue.GET("/:id", port.Queue().GetQueue)
		queue.PUT("/:id", port.Queue().UpdateQueue)
		queue.DELETE("/:id", port.Queue().DeleteQueue)

		// Streaming endpoints
		queue.POST("/monitor", port.Queue().StartStreamingAll)           // Start all
		queue.POST("/monitor/:name", port.Queue().StartStreamingByName)  // Start by name
		queue.DELETE("/monitor", port.Queue().StopStreamingAll)          // Stop all
		queue.DELETE("/monitor/:name", port.Queue().StopStreamingByName) // Stop by name
	}

	// Queue WebSocket (supports ?name=queue-name filter)
	app.GET("/ws/queues", port.Queue().HandleWebSocket)

	// Interface Monitoring
	iface := app.Group("/interfaces")
	iface.Use(port.Middleware().UserAuth())
	{
		iface.POST("/monitor", port.Interface().StartMonitoring)           // Start all
		iface.POST("/monitor/:name", port.Interface().StartMonitoringByName) // Start by name
		iface.DELETE("/monitor", port.Interface().StopMonitoring)          // Stop all
		iface.DELETE("/monitor/:name", port.Interface().StopMonitoringByName) // Stop by name
	}

	// Interface WebSocket (supports ?name=interface-name filter)
	app.GET("/ws/interfaces", port.Interface().HandleWebSocket)

	// IP Pool Management
	ippool := app.Group("/ip-pools")
	ippool.Use(port.Middleware().UserAuth())
	{
		ippool.POST("", port.IpPool().CreateIpPool)
		ippool.GET("", port.IpPool().ListIpPools)
		ippool.GET("/:id", port.IpPool().GetIpPool)
		ippool.PUT("/:id", port.IpPool().UpdateIpPool)
		ippool.DELETE("/:id", port.IpPool().DeleteIpPool)
	}

	// Bandwidth Profile Management
	bandwidthProfile := app.Group("/bandwidth-profiles")
	bandwidthProfile.Use(port.Middleware().UserAuth())
	{
		bandwidthProfile.POST("", port.BandwidthProfile().Create)
		bandwidthProfile.GET("", port.BandwidthProfile().List)
		bandwidthProfile.GET("/:id", port.BandwidthProfile().Get)
		bandwidthProfile.PUT("/:id", port.BandwidthProfile().Update)
		bandwidthProfile.DELETE("/:id", port.BandwidthProfile().Delete)
	}

	// Customer Management
	customer := app.Group("/customers")
	customer.Use(port.Middleware().UserAuth())
	{
		customer.POST("", port.Customer().Create)
		customer.GET("", port.Customer().List)
		customer.GET("/:id", port.Customer().Get)
		customer.PUT("/:id", port.Customer().Update)
		customer.DELETE("/:id", port.Customer().Delete)
		customer.GET("/:id/billing-info", port.Customer().GetBillingInfo)
		customer.POST("/:id/isolate", port.Customer().Isolate)
		customer.POST("/:id/activate", port.Customer().Activate)
	}

	// System Settings Management
	settings := app.Group("/settings")
	settings.Use(port.Middleware().UserAuth())
	{
		settings.GET("", port.SystemSetting().GetAll)
		settings.GET("/:key", port.SystemSetting().GetByKey)
		settings.PUT("/:key", port.SystemSetting().UpdateByKey)
	}

	// Public System Settings (no auth)
	app.GET("/public/settings", port.SystemSetting().GetPublic)
}
