package gin_inbound_adapter

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"go-template/internal/domain"
	"go-template/internal/model"
)

type pingAdapter struct {
	domain domain.Domain
}

func NewPingAdapter(domain domain.Domain) *pingAdapter {
	return &pingAdapter{
		domain: domain,
	}
}

func (h *pingAdapter) GetResource(c *gin.Context) {
	// TODO: Implement resource retrieval
	c.JSON(http.StatusOK, gin.H{"message": "Ping resource"})
}

func (h *pingAdapter) StartPing(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req model.PingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	if err := h.domain.Ping().StartPing(ctx, routerID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Ping started"})
}

func (h *pingAdapter) StopPing(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	address := c.Query("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "address query parameter required"})
		return
	}

	if err := h.domain.Ping().StopPing(routerID, address); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Ping stopped"})
}

func (h *pingAdapter) HandleWebSocket(c *gin.Context) {
	address := c.Query("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "address query parameter required"})
		return
	}

	// Upgrade to WebSocket
	// Note: This should be implemented similar to pppoe_websocket.go
	// For now, return not implemented
	c.JSON(http.StatusNotImplemented, gin.H{"error": "WebSocket not implemented yet"})
}
