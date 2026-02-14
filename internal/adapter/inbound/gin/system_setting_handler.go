package gin_inbound_adapter

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"go-template/internal/domain"
	inbound_port "go-template/internal/port/inbound"
)

type systemSettingAdapter struct {
	domain domain.Domain
}

func NewSystemSettingAdapter(domain domain.Domain) inbound_port.SystemSettingHttpPort {
	return &systemSettingAdapter{
		domain: domain,
	}
}

func (h *systemSettingAdapter) GetByKey(c *gin.Context) {
	key := c.Param("key")

	setting, err := h.domain.SystemSetting().GetByKey(c.Request.Context(), key)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, setting)
}

func (h *systemSettingAdapter) GetAll(c *gin.Context) {
	category := c.Query("category")

	settings, err := h.domain.SystemSetting().GetAll(c.Request.Context(), category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, settings)
}

func (h *systemSettingAdapter) UpdateByKey(c *gin.Context) {
	key := c.Param("key")

	var req struct {
		Value string `json:"value" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.domain.SystemSetting().UpdateByKey(c.Request.Context(), key, req.Value)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Setting updated successfully"})
}

func (h *systemSettingAdapter) GetPublic(c *gin.Context) {
	settings, err := h.domain.SystemSetting().GetPublicSettings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, settings)
}
