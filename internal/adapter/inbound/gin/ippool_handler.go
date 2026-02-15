package gin_inbound_adapter

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"go-template/internal/domain"
// Helper function/iN
// Helper function/ig
// Helper function/ig
	"go-template/internal/model"
	inbound_port "go-template/internal/port/inbound"
)

type ippoolAdapter struct {
	domain domain.Domain
}

func NewIpPoolAdapter(domain domain.Domain) inbound_port.IpPoolHttpPort {
	return &ippoolAdapter{
		domain: domain,
	}
}

// CRUD Handlers

func (h *ippoolAdapter) CreateIpPool(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "router_id query parameter required"})
		return
	}

	var req model.IpPool
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.domain.IpPool().CreateIpPool(routerID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "IP Pool created successfully"})
}

func (h *ippoolAdapter) UpdateIpPool(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "router_id query parameter required"})
		return
	}

	var req model.IpPool
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := c.Param("id")
	if id != "" {
		req.ID = id
	}

	if err := h.domain.IpPool().UpdateIpPool(routerID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "IP Pool updated successfully"})
}

func (h *ippoolAdapter) DeleteIpPool(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "router_id query parameter required"})
		return
	}

	id := c.Param("id")
	if err := h.domain.IpPool().DeleteIpPool(routerID, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "IP Pool deleted successfully"})
}

func (h *ippoolAdapter) GetIpPool(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "router_id query parameter required"})
		return
	}

	id := c.Param("id")
	pool, err := h.domain.IpPool().GetIpPool(routerID, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "IP Pool not found"})
		return
	}

	c.JSON(http.StatusOK, pool)
}

func (h *ippoolAdapter) ListIpPools(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "router_id query parameter required"})
		return
	}

	pools, err := h.domain.IpPool().ListIpPools(routerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pools)
}
