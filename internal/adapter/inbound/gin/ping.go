package gin_inbound_adapter

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"go-template/internal/domain"
	"go-template/internal/model"
	inbound_port "go-template/internal/port/inbound"
)

type pingAdapter struct {
	domain domain.Domain
}

func NewPingAdapter(
	domain domain.Domain,
) inbound_port.PingHttpPort {
	return &pingAdapter{
		domain: domain,
	}
}

func (h *pingAdapter) GetResource(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

// Ping Operations

func (h *pingAdapter) StartPing(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "router_id query parameter required"})
		return
	}

	var req model.PingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create a background context for the ping
	ctx := context.Background()

	if err := h.domain.Ping().StartPing(ctx, routerID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Ping started for: " + req.Address})
}

func (h *pingAdapter) StopPing(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "router_id query parameter required"})
		return
	}

	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "address is required"})
		return
	}

	if err := h.domain.Ping().StopPing(routerID, address); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Ping stopped for: " + address})
}

// WebSocket Handler

func (h *pingAdapter) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade to WebSocket"})
		return
	}
	defer conn.Close()

	ctx := c.Request.Context()

	// Get address from query parameter
	address := c.Query("address")
	if address == "" {
		return
	}

	eventChan, err := h.domain.Ping().SubscribeToPingResults(ctx, address)
	if err != nil {
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-eventChan:
			if !ok {
				return
			}
			if err := conn.WriteJSON(msg); err != nil {
				return
			}
		}
	}
}
