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
	// ─── Public Registration (no auth) ────────────────────────────────────────
	public := app.Group("/public")
	{
		public.POST("/register", port.Registration().Submit)
	}

	// ─── Customer Portal Auth (no auth) ───────────────────────────────────────
	portalAuth := app.Group("/portal/auth")
	{
		portalAuth.POST("/login", port.CustomerPortal().Login)
		portalAuth.POST("/refresh", port.CustomerPortal().RefreshToken)
	}

	// ─── Customer Portal (CustomerPortalAuth) ─────────────────────────────────
	portalMe := app.Group("/portal")
	portalMe.Use(port.Middleware().CustomerPortalAuth())
	{
		portalMe.GET("/me", port.CustomerPortal().GetProfile)
		portalMe.PUT("/me", port.CustomerPortal().UpdateProfile)
		portalMe.POST("/change-password", port.CustomerPortal().ChangePassword)
		portalMe.PUT("/ppp-credentials", port.CustomerPortal().ChangePppCredentials)
		portalMe.GET("/invoices", port.CustomerPortal().ListInvoices)
		portalMe.GET("/invoices/:id", port.CustomerPortal().GetInvoice)
	}

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
	pingMgmt.Use(port.Middleware().RBAC())
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

	// User routes (all protected by RBAC)
	user := app.Group("/user")
	user.Use(port.Middleware().UserAuth())
	user.Use(port.Middleware().RBAC())
	{
		user.POST("/change-password", port.Auth().ChangePassword)
		user.POST("/logout", port.Auth().Logout)
		user.GET("/profile", port.User().GetProfile)
		user.PUT("/profile", port.User().UpdateProfile)
	}

	// Registration Management (admin/user review pending registrations)
	registrations := app.Group("/registrations")
	registrations.Use(port.Middleware().UserAuth())
	registrations.Use(port.Middleware().RBAC())
	{
		registrations.GET("", port.Registration().List)
		registrations.GET("/:id", port.Registration().GetByID)
		registrations.POST("/:id/approve", port.Registration().Approve)
		registrations.POST("/:id/reject", port.Registration().Reject)
	}

	// PPPoE Management
	pppoe := app.Group("/pppoe")
	pppoe.Use(port.Middleware().UserAuth())
	pppoe.Use(port.Middleware().RBAC())
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
	queue.Use(port.Middleware().RBAC())
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
	iface.Use(port.Middleware().RBAC())
	{
		iface.POST("/monitor", port.Interface().StartMonitoring)              // Start all
		iface.POST("/monitor/:name", port.Interface().StartMonitoringByName)  // Start by name
		iface.DELETE("/monitor", port.Interface().StopMonitoring)             // Stop all
		iface.DELETE("/monitor/:name", port.Interface().StopMonitoringByName) // Stop by name
	}

	// Interface WebSocket (supports ?name=interface-name filter)
	app.GET("/ws/interfaces", port.Interface().HandleWebSocket)

	// IP Pool Management
	ippool := app.Group("/ip-pools")
	ippool.Use(port.Middleware().UserAuth())
	ippool.Use(port.Middleware().RBAC())
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
	bandwidthProfile.Use(port.Middleware().RBAC())
	{
		bandwidthProfile.POST("", port.BandwidthProfile().Create)
		bandwidthProfile.GET("", port.BandwidthProfile().List)
		bandwidthProfile.GET("/code/:code", port.BandwidthProfile().GetByCode)
		bandwidthProfile.GET("/:id", port.BandwidthProfile().GetByID)
		bandwidthProfile.PUT("/:id", port.BandwidthProfile().Update)
		bandwidthProfile.DELETE("/:id", port.BandwidthProfile().Delete)
		bandwidthProfile.POST("/:id/sync", port.BandwidthProfile().SyncToMikrotik)
	}

	// Customer Management (read-only, no MikroTik required)
	customer := app.Group("/customers")
	customer.Use(port.Middleware().UserAuth())
	customer.Use(port.Middleware().RBAC())
	{
		customer.GET("", port.Customer().List)
		customer.GET("/code/:code", port.Customer().GetByCode)
		customer.GET("/:id", port.Customer().GetByID)
	}

	// Invoice Management
	invoice := app.Group("/invoices")
	invoice.Use(port.Middleware().UserAuth())
	invoice.Use(port.Middleware().RBAC())
	{
		invoice.POST("", port.Invoice().Create)
		invoice.GET("", port.Invoice().List)
		invoice.GET("/number/:number", port.Invoice().GetByNumber)
		invoice.GET("/:id", port.Invoice().GetByID)
		invoice.PUT("/:id", port.Invoice().Update)
		invoice.DELETE("/:id", port.Invoice().Delete)
		invoice.POST("/generate/:customer_id", port.Invoice().GenerateMonthly)
		invoice.POST("/:id/late-fee", port.Invoice().CalculateLateFee)
	}

	// Payment Management
	payment := app.Group("/payments")
	payment.Use(port.Middleware().UserAuth())
	payment.Use(port.Middleware().RBAC())
	{
		payment.POST("", port.Payment().Create)
		payment.GET("", port.Payment().List)
		payment.GET("/number/:number", port.Payment().GetByNumber)
		payment.GET("/:id", port.Payment().GetByID)
		payment.PUT("/:id", port.Payment().Update)
		payment.DELETE("/:id", port.Payment().Delete)
		payment.POST("/:id/confirm", port.Payment().Confirm)
		payment.POST("/:id/reject", port.Payment().Reject)
		payment.POST("/:id/allocate", port.Payment().Allocate)
	}

	// ─────────────────────────────────────────────────────────────────────────
	// MikroTik Router Management
	// ─────────────────────────────────────────────────────────────────────────

	// CRUD routers (no live connection required)
	mikrotik := app.Group("/mikrotik")
	mikrotik.Use(port.Middleware().UserAuth())
	mikrotik.Use(port.Middleware().RBAC())
	{
		mikrotik.POST("", port.MikrotikRouter().Create)
		mikrotik.GET("", port.MikrotikRouter().List)
		mikrotik.GET("/:router_id", port.MikrotikRouter().GetByID)
		mikrotik.PUT("/:router_id", port.MikrotikRouter().Update)
		mikrotik.DELETE("/:router_id", port.MikrotikRouter().Delete)
		mikrotik.POST("/:router_id/test", port.MikrotikRouter().TestConnection)
	}

	// Routes requiring a live MikroTik connection — router_id resolved via RouterAuth middleware
	mikrotikOp := app.Group("/mikrotik/:router_id")
	mikrotikOp.Use(port.Middleware().UserAuth())
	mikrotikOp.Use(port.Middleware().RBAC())
	mikrotikOp.Use(port.Middleware().RouterAuth())
	{
		// PPPoE Secrets via router path
		mikrotikOp.POST("/pppoe/secrets", port.Pppoe().CreateSecret)
		mikrotikOp.GET("/pppoe/secrets", port.Pppoe().ListSecrets)
		mikrotikOp.GET("/pppoe/secrets/:id", port.Pppoe().GetSecret)
		mikrotikOp.PUT("/pppoe/secrets/:id", port.Pppoe().UpdateSecret)
		mikrotikOp.DELETE("/pppoe/secrets/:id", port.Pppoe().DeleteSecret)

		// PPPoE Profiles via router path
		mikrotikOp.POST("/pppoe/profiles", port.Pppoe().CreateProfile)
		mikrotikOp.GET("/pppoe/profiles", port.Pppoe().ListProfiles)
		mikrotikOp.GET("/pppoe/profiles/:id", port.Pppoe().GetProfile)
		mikrotikOp.PUT("/pppoe/profiles/:id", port.Pppoe().UpdateProfile)
		mikrotikOp.DELETE("/pppoe/profiles/:id", port.Pppoe().DeleteProfile)

		// PPPoE Sessions via router path
		mikrotikOp.GET("/pppoe/sessions/active", port.Pppoe().ListActiveSessions)
		mikrotikOp.GET("/pppoe/sessions/inactive", port.Pppoe().ListInactiveSessions)

		// Queues via router path
		mikrotikOp.POST("/queues", port.Queue().CreateQueue)
		mikrotikOp.GET("/queues", port.Queue().ListQueues)
		mikrotikOp.GET("/queues/:id", port.Queue().GetQueue)
		mikrotikOp.PUT("/queues/:id", port.Queue().UpdateQueue)
		mikrotikOp.DELETE("/queues/:id", port.Queue().DeleteQueue)

		// Queue monitoring via router path
		mikrotikOp.POST("/queues/monitor", port.Queue().StartStreamingAll)
		mikrotikOp.POST("/queues/monitor/:name", port.Queue().StartStreamingByName)
		mikrotikOp.DELETE("/queues/monitor", port.Queue().StopStreamingAll)
		mikrotikOp.DELETE("/queues/monitor/:name", port.Queue().StopStreamingByName)

		// Interface monitoring via router path
		mikrotikOp.POST("/interfaces/monitor", port.Interface().StartMonitoring)
		mikrotikOp.POST("/interfaces/monitor/:name", port.Interface().StartMonitoringByName)
		mikrotikOp.DELETE("/interfaces/monitor", port.Interface().StopMonitoring)
		mikrotikOp.DELETE("/interfaces/monitor/:name", port.Interface().StopMonitoringByName)

		// IP Pools via router path
		mikrotikOp.POST("/ip-pools", port.IpPool().CreateIpPool)
		mikrotikOp.GET("/ip-pools", port.IpPool().ListIpPools)
		mikrotikOp.GET("/ip-pools/:id", port.IpPool().GetIpPool)
		mikrotikOp.PUT("/ip-pools/:id", port.IpPool().UpdateIpPool)
		mikrotikOp.DELETE("/ip-pools/:id", port.IpPool().DeleteIpPool)

		// Ping via router path
		mikrotikOp.POST("/ping", port.Ping().StartPing)
		mikrotikOp.DELETE("/ping/:address", port.Ping().StopPing)

		// Customer Management (MikroTik-first: creates PPP secret on router)
		// GET routes auto-filter by router_id from path
		mikrotikOp.GET("/customers", port.Customer().List)
		mikrotikOp.GET("/customers/code/:code", port.Customer().GetByCode)
		mikrotikOp.GET("/customers/:id", port.Customer().GetByID)
		mikrotikOp.POST("/customers", port.Customer().Create)
		mikrotikOp.PUT("/customers/:id", port.Customer().Update)
		mikrotikOp.DELETE("/customers/:id", port.Customer().Delete)
		mikrotikOp.POST("/customers/:id/status", port.Customer().ChangeStatus)
		mikrotikOp.POST("/customers/:id/isolate", port.Customer().Isolate)
		mikrotikOp.POST("/customers/:id/unisolate", port.Customer().UnIsolate)
		mikrotikOp.POST("/customers/:id/sync", port.Customer().SyncToMikrotik)

		// Bandwidth Profile via router path (MikroTik-first: creates PPP profile on router)
		// GET routes list all profiles (profiles are global, not per-router)
		mikrotikOp.GET("/bandwidth-profiles", port.BandwidthProfile().List)
		mikrotikOp.GET("/bandwidth-profiles/code/:code", port.BandwidthProfile().GetByCode)
		mikrotikOp.GET("/bandwidth-profiles/:id", port.BandwidthProfile().GetByID)
		mikrotikOp.POST("/bandwidth-profiles", port.BandwidthProfile().CreateWithRouter)
		mikrotikOp.PUT("/bandwidth-profiles/:id", port.BandwidthProfile().UpdateWithRouter)
		mikrotikOp.DELETE("/bandwidth-profiles/:id", port.BandwidthProfile().DeleteWithRouter)
		mikrotikOp.POST("/bandwidth-profiles/:id/sync", port.BandwidthProfile().SyncToMikrotik)

		// Isolation Management
		mikrotikOp.POST("/isolation/setup", port.MikrotikRouter().SetupIsolation)
		mikrotikOp.GET("/isolation/status", port.MikrotikRouter().CheckIsolationSetup)

		// Hotspot Management
		// Profiles
		mikrotikOp.POST("/hotspot/profiles", port.Hotspot().CreateProfile)
		mikrotikOp.GET("/hotspot/profiles", port.Hotspot().ListProfiles)
		mikrotikOp.GET("/hotspot/profiles/:name", port.Hotspot().GetProfile)
		mikrotikOp.PUT("/hotspot/profiles/:name", port.Hotspot().UpdateProfile)
		mikrotikOp.DELETE("/hotspot/profiles/:name", port.Hotspot().DeleteProfile)

		// Users
		mikrotikOp.POST("/hotspot/users", port.Hotspot().CreateUser)
		mikrotikOp.GET("/hotspot/users", port.Hotspot().ListUsers)
		mikrotikOp.GET("/hotspot/users/:username", port.Hotspot().GetUser)
		mikrotikOp.PUT("/hotspot/users/:username", port.Hotspot().UpdateUser)
		mikrotikOp.DELETE("/hotspot/users/:username", port.Hotspot().DeleteUser)

		// Voucher Generation
		mikrotikOp.POST("/hotspot/vouchers", port.Hotspot().GenerateVouchers)

		// Sessions
		mikrotikOp.GET("/hotspot/sessions", port.Hotspot().GetActiveSessions)
		mikrotikOp.GET("/hotspot/sessions/stats", port.Hotspot().GetSessionStats)
		mikrotikOp.DELETE("/hotspot/sessions/:username", port.Hotspot().DisconnectUser)

		// Sales
		mikrotikOp.POST("/hotspot/sales", port.Hotspot().RecordSale)
		mikrotikOp.GET("/hotspot/sales", port.Hotspot().GetSales)
		mikrotikOp.GET("/hotspot/sales/revenue", port.Hotspot().GetTotalRevenue)

		// Expiry Schedulers
		mikrotikOp.POST("/hotspot/schedulers/:profile", port.Hotspot().CreateExpiryScheduler)
		mikrotikOp.DELETE("/hotspot/schedulers/:profile", port.Hotspot().RemoveExpiryScheduler)
	}
}
