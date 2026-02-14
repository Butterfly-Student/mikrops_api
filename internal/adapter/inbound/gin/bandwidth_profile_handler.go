package gin_inbound_adapter

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"go-template/internal/domain"
	"go-template/internal/model"
	inbound_port "go-template/internal/port/inbound"
)

type bandwidthProfileAdapter struct {
	domain domain.Domain
}

func NewBandwidthProfileAdapter(domain domain.Domain) inbound_port.BandwidthProfileHttpPort {
	return &bandwidthProfileAdapter{
		domain: domain,
	}
}

func (h *bandwidthProfileAdapter) Create(c *gin.Context) {
	var req model.BandwidthProfileInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	profile, err := h.domain.BandwidthProfile().Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, profile)
}

func (h *bandwidthProfileAdapter) Get(c *gin.Context) {
	id := c.Param("id")

	profile, err := h.domain.BandwidthProfile().FindByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, profile)
}

func (h *bandwidthProfileAdapter) List(c *gin.Context) {
	var filter model.BandwidthProfileFilter

	// Parse query parameters
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	profiles, err := h.domain.BandwidthProfile().FindByFilter(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, profiles)
}

func (h *bandwidthProfileAdapter) Update(c *gin.Context) {
	id := c.Param("id")

	var req model.BandwidthProfileInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	profile, err := h.domain.BandwidthProfile().Update(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, profile)
}

func (h *bandwidthProfileAdapter) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.domain.BandwidthProfile().Delete(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Bandwidth profile deleted successfully"})
}
