package gin_inbound_adapter

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go-template/internal/domain"
	"go-template/internal/model"
	inbound_port "go-template/internal/port/inbound"
)

type ActivityHandler struct {
	domain domain.Domain
}

func NewActivityHandler(domain domain.Domain) inbound_port.ActivityHttpPort {
	return &ActivityHandler{domain: domain}
}

func (h *ActivityHandler) ListLogs(c *gin.Context) {
	var filter model.ActivityLogFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set default limit
	if filter.Limit == 0 {
		filter.Limit = 100
	}

	logs, err := h.domain.Activity().ListLogs(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count": len(logs),
		"logs":  logs,
	})
}

func (h *ActivityHandler) GetEntityHistory(c *gin.Context) {
	entityType := c.Query("entity_type")
	entityID := c.Query("entity_id")

	if entityType == "" || entityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "entity_type and entity_id are required"})
		return
	}

	logs, err := h.domain.Activity().GetEntityHistory(c.Request.Context(), entityType, entityID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count": len(logs),
		"logs":  logs,
	})
}

func (h *ActivityHandler) GetUserLogs(c *gin.Context) {
	userIDStr := c.Param("user_id")

	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	filter := model.ActivityLogFilter{
		UserIDs: []uint{uint(userID)},
		Limit:   100,
	}

	logs, err := h.domain.Activity().ListLogs(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count": len(logs),
		"logs":  logs,
	})
}
