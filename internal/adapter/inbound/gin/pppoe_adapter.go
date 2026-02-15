package gin_inbound_adapter

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"go-template/internal/domain"
	"go-template/internal/model"
)

type pppoeAdapter struct {
	domain domain.Domain
}

func NewPppoeAdapter(domain domain.Domain) *pppoeAdapter {
	return &pppoeAdapter{
		domain: domain,
	}
}

// Secret Management

func (h *pppoeAdapter) CreateSecret(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req model.PppoeSecret
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.domain.Pppoe().CreateSecret(routerID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Secret created successfully"})
}

func (h *pppoeAdapter) UpdateSecret(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req model.PppoeSecret
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := c.Param("id")
	if id != "" {
		req.ID = id
	}

	if err := h.domain.Pppoe().UpdateSecret(routerID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Secret updated successfully"})
}

func (h *pppoeAdapter) DeleteSecret(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := c.Param("id")
	if err := h.domain.Pppoe().DeleteSecret(routerID, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Secret deleted successfully"})
}

func (h *pppoeAdapter) GetSecret(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := c.Param("id")
	secret, err := h.domain.Pppoe().GetSecret(routerID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, secret)
}

func (h *pppoeAdapter) ListSecrets(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	secrets, err := h.domain.Pppoe().ListSecrets(routerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, secrets)
}

// Profile Management

func (h *pppoeAdapter) CreateProfile(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req model.PppoeProfile
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.domain.Pppoe().CreateProfile(routerID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Profile created successfully"})
}

func (h *pppoeAdapter) UpdateProfile(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req model.PppoeProfile
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := c.Param("id")
	if id != "" {
		req.ID = id
	}

	if err := h.domain.Pppoe().UpdateProfile(routerID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}

func (h *pppoeAdapter) DeleteProfile(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := c.Param("id")
	if err := h.domain.Pppoe().DeleteProfile(routerID, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile deleted successfully"})
}

func (h *pppoeAdapter) GetProfile(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := c.Param("id")
	profile, err := h.domain.Pppoe().GetProfile(routerID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, profile)
}

func (h *pppoeAdapter) ListProfiles(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	profiles, err := h.domain.Pppoe().ListProfiles(routerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, profiles)
}

// Session Management

func (h *pppoeAdapter) ListActiveSessions(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sessions, err := h.domain.Pppoe().ListActiveSessions(routerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sessions)
}

func (h *pppoeAdapter) ListInactiveSessions(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sessions, err := h.domain.Pppoe().ListInactiveSessions(routerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sessions)
}

// Session History

func (h *pppoeAdapter) ListSessionHistory(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Not implemented in domain yet
	// This is a placeholder for future session history functionality
	c.JSON(http.StatusNotImplemented, gin.H{
		"message":  "Session history not implemented yet",
		"routerID": routerID,
	})
}

// Webhooks

func (h *pppoeAdapter) CallbackOnUp(c *gin.Context) {
	var callback model.PppoeCallbackData
	if err := c.ShouldBindJSON(&callback); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.domain.Pppoe().HandleOnUp(callback); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Callback processed"})
}

func (h *pppoeAdapter) CallbackOnDown(c *gin.Context) {
	var callback model.PppoeCallbackData
	if err := c.ShouldBindJSON(&callback); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.domain.Pppoe().HandleOnDown(callback); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Callback processed"})
}

// WebSocket is handled by pppoe_websocket.go
