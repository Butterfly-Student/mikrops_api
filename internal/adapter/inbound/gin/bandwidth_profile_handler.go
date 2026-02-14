package gin_inbound_adapter

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go-template/internal/domain"
	"go-template/internal/model"
	inbound_port "go-template/internal/port/inbound"
)

type BandwidthProfileHandler struct {
	domain domain.Domain
}

func NewBandwidthProfileHandler(domain domain.Domain) inbound_port.BandwidthProfileHttpPort {
	return &BandwidthProfileHandler{domain: domain}
}

func (h *BandwidthProfileHandler) CreateProfile(c *gin.Context) {
	var input model.BandwidthProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	profile, err := h.domain.BandwidthProfile().CreateProfile(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, profile)
}

func (h *BandwidthProfileHandler) GetProfile(c *gin.Context) {
	id := c.Param("id")

	profile, err := h.domain.BandwidthProfile().GetProfile(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, profile)
}

func (h *BandwidthProfileHandler) ListProfiles(c *gin.Context) {
	var filter model.BandwidthProfileFilter

	if categories := c.QueryArray("category"); len(categories) > 0 {
		filter.Categories = categories
	}
	if isActive := c.Query("is_active"); isActive != "" {
		isActiveBool := isActive == "true"
		filter.IsActive = &isActiveBool
	}
	if isVisible := c.Query("is_visible"); isVisible != "" {
		isVisibleBool := isVisible == "true"
		filter.IsVisible = &isVisibleBool
	}

	profiles, err := h.domain.BandwidthProfile().ListProfiles(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, profiles)
}

func (h *BandwidthProfileHandler) UpdateProfile(c *gin.Context) {
	id := c.Param("id")

	var input model.BandwidthProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	profile, err := h.domain.BandwidthProfile().UpdateProfile(c.Request.Context(), id, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, profile)
}

func (h *BandwidthProfileHandler) DeleteProfile(c *gin.Context) {
	id := c.Param("id")

	err := h.domain.BandwidthProfile().DeleteProfile(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile deleted successfully"})
}

func (h *BandwidthProfileHandler) SyncToMikrotik(c *gin.Context) {
	id := c.Param("id")

	profile, err := h.domain.BandwidthProfile().GetProfile(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	err = h.domain.BandwidthProfile().SyncToMikrotik(c.Request.Context(), profile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile synced to Mikrotik successfully"})
}

func (h *BandwidthProfileHandler) GetIsolatedProfile(c *gin.Context) {
	profile, err := h.domain.BandwidthProfile().GetIsolatedProfile(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, profile)
}
