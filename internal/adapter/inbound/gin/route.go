package gin_inbound_adapter

import (
	"context"

	"github.com/gin-gonic/gin"

	inbound_port "mikrops/internal/port/inbound"
)

func InitRoute(
	ctx context.Context,
	router *gin.Engine,
	port inbound_port.HttpPort,
) {
	// ========== Auth (no middleware) ==========
	auth := router.Group("/auth")
	auth.POST("/staff/login", func(c *gin.Context) { port.Auth().StaffLogin(c) })
	auth.POST("/staff/refresh", func(c *gin.Context) { port.Auth().StaffRefresh(c) })
	auth.POST("/customer/login", func(c *gin.Context) { port.Auth().CustomerLogin(c) })
	auth.POST("/customer/refresh", func(c *gin.Context) { port.Auth().CustomerRefresh(c) })

	// ========== Internal (InternalAuth) ==========
	internal := router.Group("/internal")
	internal.Use(func(c *gin.Context) { port.Middleware().InternalAuth(c) })
	internal.POST("/client-upsert", func(c *gin.Context) { port.Client().Upsert(c) })
	internal.POST("/client-find", func(c *gin.Context) { port.Client().Find(c) })
	internal.DELETE("/client-delete", func(c *gin.Context) { port.Client().Delete(c) })

	// ========== Client Auth (existing) ==========
	client := router.Group("/v1")
	client.Use(func(c *gin.Context) { port.Middleware().ClientAuth(c) })
	client.GET("/ping", func(c *gin.Context) { port.Ping().GetResource(c) })

	// ========== Staff Admin API (StaffAuth + RequirePermission) ==========
	api := router.Group("/api/v1")
	api.Use(func(c *gin.Context) { port.Middleware().StaffAuth(c) })

	// Tenant
	api.GET("/tenant", func(c *gin.Context) { port.Tenant().Get(c) })
	api.PUT("/tenant", func(c *gin.Context) {
		port.Middleware().RequirePermission("tenant", "manage")(c)
		if !c.IsAborted() {
			port.Tenant().Update(c)
		}
	})

	// Staffs
	staffs := api.Group("/staffs")
	staffs.Use(func(c *gin.Context) { port.Middleware().RequirePermission("staff", "read")(c) })
	staffs.GET("", func(c *gin.Context) { port.Staff().List(c) })
	staffs.GET("/:id", func(c *gin.Context) { port.Staff().Get(c) })
	staffs.POST("", func(c *gin.Context) {
		port.Middleware().RequirePermission("staff", "create")(c)
		if !c.IsAborted() {
			port.Staff().Create(c)
		}
	})
	staffs.PUT("/:id", func(c *gin.Context) {
		port.Middleware().RequirePermission("staff", "update")(c)
		if !c.IsAborted() {
			port.Staff().Update(c)
		}
	})
	staffs.DELETE("/:id", func(c *gin.Context) {
		port.Middleware().RequirePermission("staff", "delete")(c)
		if !c.IsAborted() {
			port.Staff().Delete(c)
		}
	})

	// NAS
	nas := api.Group("/nas")
	nas.Use(func(c *gin.Context) { port.Middleware().RequirePermission("nas", "read")(c) })
	nas.GET("", func(c *gin.Context) { port.Nas().List(c) })
	nas.GET("/:id", func(c *gin.Context) { port.Nas().Get(c) })
	nas.POST("", func(c *gin.Context) {
		port.Middleware().RequirePermission("nas", "create")(c)
		if !c.IsAborted() {
			port.Nas().Create(c)
		}
	})
	nas.PUT("/:id", func(c *gin.Context) {
		port.Middleware().RequirePermission("nas", "update")(c)
		if !c.IsAborted() {
			port.Nas().Update(c)
		}
	})
	nas.DELETE("/:id", func(c *gin.Context) {
		port.Middleware().RequirePermission("nas", "delete")(c)
		if !c.IsAborted() {
			port.Nas().Delete(c)
		}
	})
	nas.POST("/:id/test", func(c *gin.Context) { port.Nas().TestConnection(c) })

	// MikroTik Operations via NAS
	nas.GET("/:id/pppoe/secrets", func(c *gin.Context) { port.Mikrotik().ListPPPoESecrets(c) })
	nas.POST("/:id/pppoe/secrets", func(c *gin.Context) { port.Mikrotik().CreatePPPoESecret(c) })
	nas.PUT("/:id/pppoe/secrets/:sid", func(c *gin.Context) { port.Mikrotik().UpdatePPPoESecret(c) })
	nas.DELETE("/:id/pppoe/secrets/:sid", func(c *gin.Context) { port.Mikrotik().DeletePPPoESecret(c) })
	nas.POST("/:id/pppoe/secrets/:sid/disable", func(c *gin.Context) { port.Mikrotik().DisablePPPoESecret(c) })
	nas.POST("/:id/pppoe/secrets/:sid/enable", func(c *gin.Context) { port.Mikrotik().EnablePPPoESecret(c) })
	nas.GET("/:id/pppoe/profiles", func(c *gin.Context) { port.Mikrotik().ListPPPoEProfiles(c) })
	nas.POST("/:id/pppoe/profiles", func(c *gin.Context) { port.Mikrotik().CreatePPPoEProfile(c) })
	nas.GET("/:id/hotspot/users", func(c *gin.Context) { port.Mikrotik().ListHotspotUsers(c) })
	nas.POST("/:id/hotspot/users", func(c *gin.Context) { port.Mikrotik().CreateHotspotUser(c) })
	nas.PUT("/:id/hotspot/users/:uid", func(c *gin.Context) { port.Mikrotik().UpdateHotspotUser(c) })
	nas.DELETE("/:id/hotspot/users/:uid", func(c *gin.Context) { port.Mikrotik().DeleteHotspotUser(c) })
	nas.GET("/:id/queues", func(c *gin.Context) { port.Mikrotik().ListSimpleQueues(c) })
	nas.POST("/:id/queues", func(c *gin.Context) { port.Mikrotik().CreateSimpleQueue(c) })
	nas.PUT("/:id/queues/:qid", func(c *gin.Context) { port.Mikrotik().UpdateSimpleQueue(c) })
	nas.DELETE("/:id/queues/:qid", func(c *gin.Context) { port.Mikrotik().DeleteSimpleQueue(c) })
	nas.GET("/:id/connections", func(c *gin.Context) { port.Mikrotik().GetActiveConnections(c) })
	nas.GET("/:id/traffic/:interface", func(c *gin.Context) { port.Mikrotik().GetInterfaceTraffic(c) })

	// Packages
	packages := api.Group("/packages")
	packages.Use(func(c *gin.Context) { port.Middleware().RequirePermission("internet_package", "read")(c) })
	packages.GET("", func(c *gin.Context) { port.InternetPackage().List(c) })
	packages.GET("/:id", func(c *gin.Context) { port.InternetPackage().Get(c) })
	packages.POST("", func(c *gin.Context) {
		port.Middleware().RequirePermission("internet_package", "create")(c)
		if !c.IsAborted() {
			port.InternetPackage().Create(c)
		}
	})
	packages.PUT("/:id", func(c *gin.Context) {
		port.Middleware().RequirePermission("internet_package", "update")(c)
		if !c.IsAborted() {
			port.InternetPackage().Update(c)
		}
	})
	packages.DELETE("/:id", func(c *gin.Context) {
		port.Middleware().RequirePermission("internet_package", "delete")(c)
		if !c.IsAborted() {
			port.InternetPackage().Delete(c)
		}
	})

	// Customers
	customers := api.Group("/customers")
	customers.Use(func(c *gin.Context) { port.Middleware().RequirePermission("customer", "read")(c) })
	customers.GET("", func(c *gin.Context) { port.Customer().List(c) })
	customers.GET("/:id", func(c *gin.Context) { port.Customer().Get(c) })
	customers.POST("", func(c *gin.Context) {
		port.Middleware().RequirePermission("customer", "create")(c)
		if !c.IsAborted() {
			port.Customer().Create(c)
		}
	})
	customers.PUT("/:id", func(c *gin.Context) {
		port.Middleware().RequirePermission("customer", "update")(c)
		if !c.IsAborted() {
			port.Customer().Update(c)
		}
	})
	customers.DELETE("/:id", func(c *gin.Context) {
		port.Middleware().RequirePermission("customer", "delete")(c)
		if !c.IsAborted() {
			port.Customer().Delete(c)
		}
	})

	// Subscriptions
	subscriptions := api.Group("/subscriptions")
	subscriptions.Use(func(c *gin.Context) { port.Middleware().RequirePermission("subscription", "read")(c) })
	subscriptions.GET("", func(c *gin.Context) { port.Subscription().List(c) })
	subscriptions.GET("/:id", func(c *gin.Context) { port.Subscription().Get(c) })
	subscriptions.POST("", func(c *gin.Context) {
		port.Middleware().RequirePermission("subscription", "create")(c)
		if !c.IsAborted() {
			port.Subscription().Create(c)
		}
	})
	subscriptions.PUT("/:id", func(c *gin.Context) {
		port.Middleware().RequirePermission("subscription", "update")(c)
		if !c.IsAborted() {
			port.Subscription().Update(c)
		}
	})
	subscriptions.POST("/:id/suspend", func(c *gin.Context) {
		port.Middleware().RequirePermission("subscription", "update")(c)
		if !c.IsAborted() {
			port.Subscription().Suspend(c)
		}
	})
	subscriptions.POST("/:id/activate", func(c *gin.Context) {
		port.Middleware().RequirePermission("subscription", "update")(c)
		if !c.IsAborted() {
			port.Subscription().Activate(c)
		}
	})
	subscriptions.POST("/:id/cancel", func(c *gin.Context) {
		port.Middleware().RequirePermission("subscription", "update")(c)
		if !c.IsAborted() {
			port.Subscription().Cancel(c)
		}
	})
	subscriptions.POST("/:id/vacation", func(c *gin.Context) {
		port.Middleware().RequirePermission("subscription", "update")(c)
		if !c.IsAborted() {
			port.Subscription().SetVacation(c)
		}
	})

	// Payment Methods
	paymentMethods := api.Group("/payment-methods")
	paymentMethods.Use(func(c *gin.Context) { port.Middleware().RequirePermission("payment_method", "read")(c) })
	paymentMethods.GET("", func(c *gin.Context) { port.PaymentMethod().List(c) })
	paymentMethods.GET("/:id", func(c *gin.Context) { port.PaymentMethod().Get(c) })
	paymentMethods.POST("", func(c *gin.Context) {
		port.Middleware().RequirePermission("payment_method", "create")(c)
		if !c.IsAborted() {
			port.PaymentMethod().Create(c)
		}
	})
	paymentMethods.PUT("/:id", func(c *gin.Context) {
		port.Middleware().RequirePermission("payment_method", "update")(c)
		if !c.IsAborted() {
			port.PaymentMethod().Update(c)
		}
	})
	paymentMethods.DELETE("/:id", func(c *gin.Context) {
		port.Middleware().RequirePermission("payment_method", "delete")(c)
		if !c.IsAborted() {
			port.PaymentMethod().Delete(c)
		}
	})

	// Invoices
	invoices := api.Group("/invoices")
	invoices.Use(func(c *gin.Context) { port.Middleware().RequirePermission("invoice", "read")(c) })
	invoices.GET("", func(c *gin.Context) { port.Invoice().List(c) })
	invoices.GET("/:id", func(c *gin.Context) { port.Invoice().Get(c) })
	invoices.POST("", func(c *gin.Context) {
		port.Middleware().RequirePermission("invoice", "create")(c)
		if !c.IsAborted() {
			port.Invoice().Create(c)
		}
	})
	invoices.PUT("/:id", func(c *gin.Context) {
		port.Middleware().RequirePermission("invoice", "update")(c)
		if !c.IsAborted() {
			port.Invoice().Update(c)
		}
	})
	invoices.POST("/generate", func(c *gin.Context) {
		port.Middleware().RequirePermission("invoice", "create")(c)
		if !c.IsAborted() {
			port.Invoice().GenerateBulk(c)
		}
	})

	// Payments
	payments := api.Group("/payments")
	payments.Use(func(c *gin.Context) { port.Middleware().RequirePermission("payment", "read")(c) })
	payments.GET("", func(c *gin.Context) { port.Payment().List(c) })
	payments.GET("/:id", func(c *gin.Context) { port.Payment().Get(c) })
	payments.POST("/:id/verify", func(c *gin.Context) {
		port.Middleware().RequirePermission("payment", "update")(c)
		if !c.IsAborted() {
			port.Payment().Verify(c)
		}
	})
	payments.POST("/:id/reject", func(c *gin.Context) {
		port.Middleware().RequirePermission("payment", "update")(c)
		if !c.IsAborted() {
			port.Payment().Reject(c)
		}
	})

	// Tenant Settings
	api.GET("/tenant/settings", func(c *gin.Context) {
		port.Middleware().RequirePermission("tenant", "manage")(c)
		if !c.IsAborted() {
			port.TenantSetting().Get(c)
		}
	})
	api.PUT("/tenant/settings", func(c *gin.Context) {
		port.Middleware().RequirePermission("tenant", "manage")(c)
		if !c.IsAborted() {
			port.TenantSetting().Upsert(c)
		}
	})

	// Customer Registrations
	registrations := api.Group("/registrations")
	registrations.Use(func(c *gin.Context) { port.Middleware().RequirePermission("customer", "read")(c) })
	registrations.GET("", func(c *gin.Context) { port.CustomerRegistration().List(c) })
	registrations.GET("/:id", func(c *gin.Context) { port.CustomerRegistration().Get(c) })
	registrations.POST("", func(c *gin.Context) {
		port.Middleware().RequirePermission("customer", "create")(c)
		if !c.IsAborted() {
			port.CustomerRegistration().Create(c)
		}
	})
	registrations.POST("/:id/approve", func(c *gin.Context) {
		port.Middleware().RequirePermission("customer", "update")(c)
		if !c.IsAborted() {
			port.CustomerRegistration().Approve(c)
		}
	})
	registrations.POST("/:id/reject", func(c *gin.Context) {
		port.Middleware().RequirePermission("customer", "update")(c)
		if !c.IsAborted() {
			port.CustomerRegistration().Reject(c)
		}
	})

	// PPPoE Accounts
	pppoeAccounts := api.Group("/pppoe-accounts")
	pppoeAccounts.Use(func(c *gin.Context) { port.Middleware().RequirePermission("customer", "read")(c) })
	pppoeAccounts.GET("", func(c *gin.Context) { port.PppoeAccount().List(c) })
	pppoeAccounts.GET("/:id", func(c *gin.Context) { port.PppoeAccount().Get(c) })
	pppoeAccounts.POST("", func(c *gin.Context) {
		port.Middleware().RequirePermission("customer", "create")(c)
		if !c.IsAborted() {
			port.PppoeAccount().Create(c)
		}
	})
	pppoeAccounts.PUT("/:id", func(c *gin.Context) {
		port.Middleware().RequirePermission("customer", "update")(c)
		if !c.IsAborted() {
			port.PppoeAccount().Update(c)
		}
	})
	pppoeAccounts.DELETE("/:id", func(c *gin.Context) {
		port.Middleware().RequirePermission("customer", "delete")(c)
		if !c.IsAborted() {
			port.PppoeAccount().Delete(c)
		}
	})
	pppoeAccounts.POST("/:id/isolate", func(c *gin.Context) {
		port.Middleware().RequirePermission("customer", "update")(c)
		if !c.IsAborted() {
			port.PppoeAccount().Isolate(c)
		}
	})
	pppoeAccounts.POST("/:id/restore", func(c *gin.Context) {
		port.Middleware().RequirePermission("customer", "update")(c)
		if !c.IsAborted() {
			port.PppoeAccount().Restore(c)
		}
	})

	// Activity Logs
	activityLogs := api.Group("/activity-logs")
	activityLogs.Use(func(c *gin.Context) { port.Middleware().RequirePermission("tenant", "manage")(c) })
	activityLogs.GET("", func(c *gin.Context) { port.ActivityLog().List(c) })
	activityLogs.GET("/:id", func(c *gin.Context) { port.ActivityLog().Get(c) })

	// MikroTik Sync Logs
	syncLogs := api.Group("/sync-logs")
	syncLogs.Use(func(c *gin.Context) { port.Middleware().RequirePermission("nas", "read")(c) })
	syncLogs.GET("", func(c *gin.Context) { port.MikrotikSyncLog().List(c) })
	syncLogs.GET("/:id", func(c *gin.Context) { port.MikrotikSyncLog().Get(c) })

	// ========== Customer Portal (CustomerAuth) ==========
	portal := router.Group("/portal/v1")
	portal.Use(func(c *gin.Context) { port.Middleware().CustomerAuth(c) })

	portal.GET("/dashboard", func(c *gin.Context) { port.Portal().Dashboard(c) })
	portal.GET("/profile", func(c *gin.Context) { port.Portal().GetProfile(c) })
	portal.PUT("/profile", func(c *gin.Context) { port.Portal().UpdateProfile(c) })
	portal.PUT("/profile/password", func(c *gin.Context) { port.Portal().UpdatePassword(c) })
	portal.GET("/subscription", func(c *gin.Context) { port.Portal().GetSubscription(c) })
	portal.GET("/connection/status", func(c *gin.Context) { port.Portal().GetConnectionStatus(c) })
	portal.GET("/connection/bandwidth", func(c *gin.Context) { port.Portal().GetBandwidth(c) })
	portal.GET("/invoices", func(c *gin.Context) { port.Portal().ListInvoices(c) })
	portal.GET("/invoices/:id", func(c *gin.Context) { port.Portal().GetInvoice(c) })
	portal.GET("/payments", func(c *gin.Context) { port.Portal().ListPayments(c) })
	portal.POST("/payments", func(c *gin.Context) { port.Portal().CreatePayment(c) })
	portal.GET("/payments/:id", func(c *gin.Context) { port.Portal().GetPayment(c) })
	portal.POST("/payments/:id/upload-proof", func(c *gin.Context) { port.Portal().UploadProof(c) })
	portal.GET("/packages", func(c *gin.Context) { port.Portal().ListPackages(c) })
	portal.GET("/payment-methods", func(c *gin.Context) { port.Portal().ListPaymentMethods(c) })
}
