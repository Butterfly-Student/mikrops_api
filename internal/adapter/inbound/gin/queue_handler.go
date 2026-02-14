package gin_inbound_adapter

import (
	"context"
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

// Enhanced Streaming with context

func (h *queueAdapter) StartStreamingAll(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "router_id query parameter required"})
		return
	}

	// Create context that will be cancelled when request is cancelled
	ctx := c.Request.Context()

	if err := h.domain.Queue().StartStreamingAll(ctx, routerID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Streaming started for all queues"})
}

func (h *queueAdapter) StartStreamingByName(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "router_id query parameter required"})
		return
	}

	queueName := c.Param("name")
	if queueName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "queue name is required"})
		return
	}

	// Create a background context for the stream
	// Note: This will run until the application stops
	// For production, consider storing context in a manager for cancellation
	ctx := context.Background()

	if err := h.domain.Queue().StartStreamingByName(ctx, routerID, queueName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Streaming started for queue: " + queueName})
}

func (h *queueAdapter) StopStreamingAll(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "router_id query parameter required"})
		return
	}

	if err := h.domain.Queue().StopStreamingAll(routerID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All queue streaming stopped"})
}

func (h *queueAdapter) StopStreamingByName(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "router_id query parameter required"})
		return
	}

	queueName := c.Param("name")
	if queueName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "queue name is required"})
		return
	}

	if err := h.domain.Queue().StopStreamingByName(routerID, queueName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Streaming stopped for queue: " + queueName})
}
