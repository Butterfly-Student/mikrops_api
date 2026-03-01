package gin_inbound_adapter

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"go-template/pkg/gowa"

	"go-template/internal/domain"
	inbound_port "go-template/internal/port/inbound"
)

type gowaAdapter struct {
	domain domain.Domain
}

func NewGowaAdapter(domain domain.Domain) inbound_port.GowaHttpPort {
	return &gowaAdapter{domain: domain}
}


// deviceIDFromQuery returns the optional device_id query param.
func deviceIDFromQuery(c *gin.Context) string {
	return c.Query("device_id")
}

// ─── Device Management ────────────────────────────────────────────────────────

func (h *gowaAdapter) ListDevices(c *gin.Context) {
	devices, err := h.domain.Gowa().ListDevices(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": devices})
}

func (h *gowaAdapter) AddDevice(c *gin.Context) {
	var req struct {
		DeviceID string `json:"device_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	device, err := h.domain.Gowa().AddDevice(c.Request.Context(), req.DeviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, device)
}

func (h *gowaAdapter) GetDevice(c *gin.Context) {
	deviceID := c.Param("device_id")
	device, err := h.domain.Gowa().GetDevice(c.Request.Context(), deviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, device)
}

func (h *gowaAdapter) RemoveDevice(c *gin.Context) {
	deviceID := c.Param("device_id")
	if err := h.domain.Gowa().RemoveDevice(c.Request.Context(), deviceID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "device removed"})
}

func (h *gowaAdapter) LoginDevice(c *gin.Context) {
	deviceID := c.Param("device_id")
	resp, err := h.domain.Gowa().LoginDevice(c.Request.Context(), deviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *gowaAdapter) LoginDeviceWithCode(c *gin.Context) {
	deviceID := c.Param("device_id")
	phone := c.Query("phone")
	if phone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "phone query param is required"})
		return
	}
	resp, err := h.domain.Gowa().LoginDeviceWithCode(c.Request.Context(), deviceID, phone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *gowaAdapter) LogoutDevice(c *gin.Context) {
	deviceID := c.Param("device_id")
	if err := h.domain.Gowa().LogoutDevice(c.Request.Context(), deviceID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "device logged out"})
}

func (h *gowaAdapter) ReconnectDevice(c *gin.Context) {
	deviceID := c.Param("device_id")
	if err := h.domain.Gowa().ReconnectDevice(c.Request.Context(), deviceID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "device reconnected"})
}

func (h *gowaAdapter) GetDeviceStatus(c *gin.Context) {
	deviceID := c.Param("device_id")
	status, err := h.domain.Gowa().GetDeviceStatus(c.Request.Context(), deviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, status)
}

// ─── App / Session ────────────────────────────────────────────────────────────

func (h *gowaAdapter) AppLogin(c *gin.Context) {
	deviceID := deviceIDFromQuery(c)
	resp, err := h.domain.Gowa().AppLogin(c.Request.Context(), deviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *gowaAdapter) AppLoginWithCode(c *gin.Context) {
	phone := c.Query("phone")
	if phone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "phone query param is required"})
		return
	}
	deviceID := deviceIDFromQuery(c)
	resp, err := h.domain.Gowa().AppLoginWithCode(c.Request.Context(), phone, deviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *gowaAdapter) AppLogout(c *gin.Context) {
	deviceID := deviceIDFromQuery(c)
	if err := h.domain.Gowa().AppLogout(c.Request.Context(), deviceID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

func (h *gowaAdapter) AppReconnect(c *gin.Context) {
	deviceID := deviceIDFromQuery(c)
	if err := h.domain.Gowa().AppReconnect(c.Request.Context(), deviceID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "reconnected"})
}

func (h *gowaAdapter) AppStatus(c *gin.Context) {
	deviceID := deviceIDFromQuery(c)
	status, err := h.domain.Gowa().AppStatus(c.Request.Context(), deviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, status)
}

// ─── Group Management ─────────────────────────────────────────────────────────

func (h *gowaAdapter) GetMyGroups(c *gin.Context) {
	deviceID := deviceIDFromQuery(c)
	groups, err := h.domain.Gowa().GetMyGroups(c.Request.Context(), deviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": groups})
}

func (h *gowaAdapter) FindGroupByName(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name query param is required"})
		return
	}
	deviceID := deviceIDFromQuery(c)
	group, err := h.domain.Gowa().FindGroupByName(c.Request.Context(), name, deviceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, group)
}

func (h *gowaAdapter) GetGroupInfo(c *gin.Context) {
	groupJID := c.Query("group_id")
	if groupJID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group_id query param is required"})
		return
	}
	deviceID := deviceIDFromQuery(c)
	info, err := h.domain.Gowa().GetGroupInfo(c.Request.Context(), groupJID, deviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, info)
}

func (h *gowaAdapter) GetGroupInviteLink(c *gin.Context) {
	groupJID := c.Query("group_id")
	if groupJID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group_id query param is required"})
		return
	}
	deviceID := deviceIDFromQuery(c)
	resp, err := h.domain.Gowa().GetGroupInviteLink(c.Request.Context(), groupJID, deviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ─── User Information ─────────────────────────────────────────────────────────

func (h *gowaAdapter) CheckUser(c *gin.Context) {
	phone := c.Query("phone")
	if phone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "phone query param is required"})
		return
	}
	deviceID := deviceIDFromQuery(c)
	isOnWhatsApp, err := h.domain.Gowa().CheckUser(c.Request.Context(), phone, deviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"phone": phone, "is_on_whatsapp": isOnWhatsApp})
}

func (h *gowaAdapter) GetUserInfo(c *gin.Context) {
	phone := c.Query("phone")
	if phone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "phone query param is required"})
		return
	}
	deviceID := deviceIDFromQuery(c)
	info, err := h.domain.Gowa().GetUserInfo(c.Request.Context(), phone, deviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, info)
}

func (h *gowaAdapter) GetMyContacts(c *gin.Context) {
	deviceID := deviceIDFromQuery(c)
	contacts, err := h.domain.Gowa().GetMyContacts(c.Request.Context(), deviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": contacts})
}

// ─── Send Messages ────────────────────────────────────────────────────────────

func (h *gowaAdapter) SendMessage(c *gin.Context) {
	var req struct {
		Phone          string   `json:"phone" binding:"required"`
		Message        string   `json:"message" binding:"required"`
		DeviceID       string   `json:"device_id"`
		ReplyMessageID string   `json:"reply_message_id"`
		IsForwarded    bool     `json:"is_forwarded"`
		Duration       int      `json:"duration"`
		Mentions       []string `json:"mentions"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sendReq := gowa.SendMessageRequest{
		Phone:          req.Phone,
		Message:        req.Message,
		ReplyMessageID: req.ReplyMessageID,
		IsForwarded:    req.IsForwarded,
		Duration:       req.Duration,
		Mentions:       req.Mentions,
	}

	resp, err := h.domain.Gowa().SendMessage(c.Request.Context(), sendReq, req.DeviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *gowaAdapter) SendImageFromURL(c *gin.Context) {
	var req struct {
		Phone    string `json:"phone" binding:"required"`
		ImageURL string `json:"image_url" binding:"required"`
		Caption  string `json:"caption"`
		DeviceID string `json:"device_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.domain.Gowa().SendImageFromURL(c.Request.Context(), req.Phone, req.ImageURL, req.Caption, req.DeviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *gowaAdapter) SendFileFromURL(c *gin.Context) {
	var req struct {
		Phone    string `json:"phone" binding:"required"`
		FileURL  string `json:"file_url" binding:"required"`
		Caption  string `json:"caption"`
		DeviceID string `json:"device_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.domain.Gowa().SendFileFromURL(c.Request.Context(), req.Phone, req.FileURL, req.Caption, req.DeviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *gowaAdapter) SendVideoFromURL(c *gin.Context) {
	var req struct {
		Phone    string `json:"phone" binding:"required"`
		VideoURL string `json:"video_url" binding:"required"`
		Caption  string `json:"caption"`
		DeviceID string `json:"device_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.domain.Gowa().SendVideoFromURL(c.Request.Context(), req.Phone, req.VideoURL, req.Caption, req.DeviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}
