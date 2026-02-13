package gin_inbound_adapter

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"go-template/internal/model"
)

// Helper to get routerID
func getRouterID(c *gin.Context) (uint, error) {
	idStr := c.Query("router_id")
	if idStr == "" {
		return 0, strconv.ErrSyntax
	}
	id, err := strconv.Atoi(idStr)
	return uint(id), err
}

// Secret Management

func (h *pppoeAdapter) CreateSecret(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "router_id query parameter required"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "router_id query parameter required"})
		return
	}

	var req model.PppoeSecret
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// ID might come from path or body, here path overrides body if present
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "router_id query parameter required"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "router_id query parameter required"})
		return
	}

	id := c.Param("id")
	secret, err := h.domain.Pppoe().GetSecret(routerID, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Secret not found"})
		return
	}

	c.JSON(http.StatusOK, secret)
}

func (h *pppoeAdapter) ListSecrets(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "router_id query parameter required"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "router_id query parameter required"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "router_id query parameter required"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "router_id query parameter required"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "router_id query parameter required"})
		return
	}

	id := c.Param("id")
	profile, err := h.domain.Pppoe().GetProfile(routerID, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}

	c.JSON(http.StatusOK, profile)
}

func (h *pppoeAdapter) ListProfiles(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "router_id query parameter required"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "router_id query parameter required"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "router_id query parameter required"})
		return
	}

	sessions, err := h.domain.Pppoe().ListInactiveSessions(routerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sessions)
}

func (h *pppoeAdapter) ListSessionHistory(c *gin.Context) {
	// Not implemented in domain per revised plan (no local history)
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}
