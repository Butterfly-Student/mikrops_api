package gin_inbound_adapter

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"go-template/internal/domain"
	"go-template/internal/model"
	inbound_port "go-template/internal/port/inbound"
)

type interfaceAdapter struct {
	domain domain.Domain
}

func NewInterfaceAdapter(domain domain.Domain) inbound_port.InterfaceHttpPort {
	return &interfaceAdapter{
		domain: domain,
	}
}

// Monitoring Handlers

func (h *interfaceAdapter) StartMonitoring(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "router_id query parameter required"})
		return
	}

	// Create a background context for the stream
	ctx := context.Background()

	if err := h.domain.Interface().StartMonitoring(ctx, routerID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Interface monitoring started for all interfaces"})
}

func (h *interfaceAdapter) StartMonitoringByName(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "router_id query parameter required"})
		return
	}

	interfaceName := c.Param("name")
	if interfaceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "interface name is required"})
		return
	}

	// Create a background context for the stream
	ctx := context.Background()

	if err := h.domain.Interface().StartMonitoringByName(ctx, routerID, interfaceName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Interface monitoring started for: " + interfaceName})
}

func (h *interfaceAdapter) StopMonitoring(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "router_id query parameter required"})
		return
	}

	if err := h.domain.Interface().StopMonitoring(routerID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Interface monitoring stopped"})
}

func (h *interfaceAdapter) StopMonitoringByName(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "router_id query parameter required"})
		return
	}

	interfaceName := c.Param("name")
	if interfaceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "interface name is required"})
		return
	}

	if err := h.domain.Interface().StopMonitoringByName(routerID, interfaceName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Interface monitoring stopped for: " + interfaceName})
}

// WebSocket Handler

func (h *interfaceAdapter) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade to WebSocket"})
		return
	}
	defer conn.Close()

	ctx := c.Request.Context()

	// Check if filtering by interface name
	interfaceName := c.Query("name")

	var eventChan <-chan model.WebSocketMessage

	if interfaceName != "" {
		// Subscribe to specific interface stats
		eventChan, err = h.domain.Interface().SubscribeToStatsByName(ctx, interfaceName)
		if err != nil {
			return
		}
	} else {
		// Subscribe to all interface stats
		eventChan, err = h.domain.Interface().SubscribeToStats(ctx)
		if err != nil {
			return
		}
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
