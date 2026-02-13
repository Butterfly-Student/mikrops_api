package gin_inbound_adapter

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"go-template/internal/model"
)

func (h *pppoeAdapter) CallbackOnUp(c *gin.Context) {
	var req model.PppoeCallbackData

	// RouterID from query or path? Assuming query
	if ridStr := c.Query("router_id"); ridStr != "" {
		rid, _ := strconv.Atoi(ridStr)
		req.RouterID = uint(rid)
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.domain.Pppoe().HandleOnUp(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func (h *pppoeAdapter) CallbackOnDown(c *gin.Context) {
	var req model.PppoeCallbackData

	if ridStr := c.Query("router_id"); ridStr != "" {
		rid, _ := strconv.Atoi(ridStr)
		req.RouterID = uint(rid)
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.domain.Pppoe().HandleOnDown(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}
