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
	internal := router.Group("/internal")
	internal.Use(func(c *gin.Context) {
		port.Middleware().InternalAuth(c)
	})
	internal.POST("/client-upsert", func(c *gin.Context) {
		port.Client().Upsert(c)
	})
	internal.POST("/client-find", func(c *gin.Context) {
		port.Client().Find(c)
	})
	internal.DELETE("/client-delete", func(c *gin.Context) {
		port.Client().Delete(c)
	})

	client := router.Group("/v1")
	client.Use(func(c *gin.Context) {
		port.Middleware().ClientAuth(c)
	})
	client.GET("/ping", func(c *gin.Context) {
		port.Ping().GetResource(c)
	})
}
