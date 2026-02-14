package gin_inbound_adapter

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go-template/internal/domain"
	"go-template/internal/model"
	inbound_port "go-template/internal/port/inbound"
)

type SystemSettingHandler struct {
	domain domain.Domain
}

func NewSystemSettingHandler(domain domain.Domain) inbound_port.SystemSettingHttpPort {
	return &SystemSettingHandler{domain: domain}
}

func (h *SystemSettingHandler) GetSetting(c *gin.Context) {
	key := c.Param("key")

	setting, err := h.domain.SystemSetting().GetSetting(c.Request.Context(), key)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, setting)
}

func (h *SystemSettingHandler) ListSettings(c *gin.Context) {
	category := c.Query("category")

	var filter model.SystemSettingFilter
	if category != "" {
		filter.Category = &category
	}

	settings, err := h.domain.SystemSetting().ListSettings(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, settings)
}

func (h *SystemSettingHandler) UpdateSetting(c *gin.Context) {
	key := c.Param("key")

	var input struct {
		Value string `json:"value" validate:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")
	err := h.domain.SystemSetting().UpdateSetting(c.Request.Context(), key, input.Value, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Setting updated successfully"})
}
