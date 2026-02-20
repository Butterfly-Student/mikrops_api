package gin_inbound_adapter

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"go-template/internal/model"
)

// Profile Management

func (h *hotspotAdapter) CreateProfile(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req model.HotspotProfile
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.domain.Hotspot().CreateProfile(c.Request.Context(), routerID, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Profile created successfully"})
}

func (h *hotspotAdapter) GetProfile(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	name := c.Param("name")
	profile, err := h.domain.Hotspot().GetProfile(c.Request.Context(), routerID, name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}

	c.JSON(http.StatusOK, profile)
}

func (h *hotspotAdapter) ListProfiles(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	profiles, err := h.domain.Hotspot().GetAllProfiles(c.Request.Context(), routerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, profiles)
}

func (h *hotspotAdapter) UpdateProfile(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	name := c.Param("name")
	var req model.HotspotProfileUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.domain.Hotspot().UpdateProfile(c.Request.Context(), routerID, name, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}

func (h *hotspotAdapter) DeleteProfile(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	name := c.Param("name")
	if err := h.domain.Hotspot().DeleteProfile(c.Request.Context(), routerID, name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile deleted successfully"})
}

// User Management

func (h *hotspotAdapter) CreateUser(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req model.HotspotUser
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.domain.Hotspot().CreateUser(c.Request.Context(), routerID, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
}

func (h *hotspotAdapter) GetUser(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username := c.Param("username")
	user, err := h.domain.Hotspot().GetUser(c.Request.Context(), routerID, username)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *hotspotAdapter) ListUsers(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	filter := &model.HotspotUserFilter{
		Profile: c.Query("profile"),
	}

	users, err := h.domain.Hotspot().GetAllUsers(c.Request.Context(), routerID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

func (h *hotspotAdapter) UpdateUser(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username := c.Param("username")
	var req model.HotspotUserUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.domain.Hotspot().UpdateUser(c.Request.Context(), routerID, username, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

func (h *hotspotAdapter) DeleteUser(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username := c.Param("username")
	if err := h.domain.Hotspot().DeleteUser(c.Request.Context(), routerID, username); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// Voucher Generation

// GenerateVouchers handles both modes via the "mode" field in the request body:
//   - "vc" (default, or username==password): voucher mode
//   - "up" (username!=password):             user-password mode
//
// Set quantity=1 for a single voucher, quantity=N for bulk generation.
func (h *hotspotAdapter) GenerateVouchers(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req model.HotspotVoucherGenerator
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	var result *model.HotspotVoucherResult

	switch req.Mode {
	case "up":
		result, err = h.domain.Hotspot().GenerateUserPasswordMode(ctx, routerID, &req)
	default:
		// "vc" is the default; any unrecognised value also falls here
		req.Mode = "vc"
		result, err = h.domain.Hotspot().GenerateVouchers(ctx, routerID, &req)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// Session Management

func (h *hotspotAdapter) GetActiveSessions(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sessions, err := h.domain.Hotspot().GetActiveSessions(c.Request.Context(), routerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sessions)
}

func (h *hotspotAdapter) GetSessionStats(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	stats, err := h.domain.Hotspot().GetSessionStats(c.Request.Context(), routerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *hotspotAdapter) DisconnectUser(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username := c.Param("username")
	if err := h.domain.Hotspot().DisconnectUser(c.Request.Context(), routerID, username); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User disconnected successfully"})
}

// Sales

func (h *hotspotAdapter) RecordSale(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req model.HotspotSale
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.domain.Hotspot().RecordSale(c.Request.Context(), routerID, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Sale recorded successfully"})
}

func (h *hotspotAdapter) GetSales(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	filter := &model.HotspotSaleFilter{
		StartDate: c.Query("start_date"),
		EndDate:   c.Query("end_date"),
		Prefix:    c.Query("prefix"),
	}

	sales, err := h.domain.Hotspot().GetAllSales(c.Request.Context(), routerID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sales)
}

func (h *hotspotAdapter) GetTotalRevenue(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	revenue, err := h.domain.Hotspot().GetTotalRevenue(c.Request.Context(), routerID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"revenue": revenue, "start_date": startDate, "end_date": endDate})
}

// Expiry Schedulers

func (h *hotspotAdapter) CreateExpiryScheduler(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	profileName := c.Param("profile")
	if err := h.domain.Hotspot().CreateExpiryScheduler(c.Request.Context(), routerID, profileName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Expiry scheduler created successfully"})
}

func (h *hotspotAdapter) RemoveExpiryScheduler(c *gin.Context) {
	routerID, err := getRouterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	profileName := c.Param("profile")
	if err := h.domain.Hotspot().RemoveExpiryScheduler(c.Request.Context(), routerID, profileName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Expiry scheduler removed successfully"})
}
