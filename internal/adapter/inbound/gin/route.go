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
}
