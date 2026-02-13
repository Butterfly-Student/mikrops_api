package gin_inbound_adapter

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"go-template/internal/domain"
	"go-template/internal/model"
	inbound_port "go-template/internal/port/inbound"
)

type queueAdapter struct {
	domain domain.Domain
}

func NewQueueAdapter(domain domain.Domain) inbound_port.QueueHttpPort {
	return &queueAdapter{
		domain: domain,
	}
}

// CRUD Handlers

func (h *queueAdapter) CreateQueue(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "router_id query parameter required"})
		return
	}

	var req model.PppoeQueue
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.domain.Queue().CreateQueue(routerID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Queue created successfully"})
}

func (h *queueAdapter) UpdateQueue(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "router_id query parameter required"})
		return
	}

	var req model.PppoeQueue
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.ID = c.Param("id")

	if err := h.domain.Queue().UpdateQueue(routerID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Queue updated successfully"})
}

func (h *queueAdapter) DeleteQueue(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "router_id query parameter required"})
		return
	}

	if err := h.domain.Queue().DeleteQueue(routerID, c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Queue deleted successfully"})
}

func (h *queueAdapter) GetQueue(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "router_id query parameter required"})
		return
	}

	queue, err := h.domain.Queue().GetQueue(routerID, c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Queue not found"})
		return
	}

	c.JSON(http.StatusOK, queue)
}

func (h *queueAdapter) ListQueues(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "router_id query parameter required"})
		return
	}

	queues, err := h.domain.Queue().ListQueues(routerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, queues)
}

// Streaming

func (h *queueAdapter) StartStreaming(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "router_id query parameter required"})
		return
	}

	if err := h.domain.Queue().StartStreaming(routerID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Streaming started"})
}
