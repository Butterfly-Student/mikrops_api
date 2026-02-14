package gin_inbound_adapter

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go-template/internal/domain"
	"go-template/internal/model"
	inbound_port "go-template/internal/port/inbound"
)

type NotificationHandler struct {
	domain domain.Domain
}

func NewNotificationHandler(domain domain.Domain) inbound_port.NotificationHttpPort {
	return &NotificationHandler{domain: domain}
}

func (h *NotificationHandler) CreateNotification(c *gin.Context) {
	var input model.NotificationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	notification, err := h.domain.Notification().CreateNotification(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, notification)
}

func (h *NotificationHandler) GetNotification(c *gin.Context) {
	id := c.Param("id")

	notification, err := h.domain.Notification().GetNotification(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, notification)
}

func (h *NotificationHandler) ListNotifications(c *gin.Context) {
	var filter model.NotificationFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	notifications, err := h.domain.Notification().ListNotifications(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count":         len(notifications),
		"notifications": notifications,
	})
}

func (h *NotificationHandler) SendNotification(c *gin.Context) {
	id := c.Param("id")

	err := h.domain.Notification().SendNotification(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "notification sent successfully"})
}

func (h *NotificationHandler) RetryFailedNotifications(c *gin.Context) {
	retried, err := h.domain.Notification().RetryFailedNotifications(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "notifications retried successfully",
		"retried":       len(retried),
		"retried_items": retried,
	})
}

func (h *NotificationHandler) CreateTemplate(c *gin.Context) {
	var input model.NotificationTemplateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	template, err := h.domain.Notification().CreateTemplate(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, template)
}

func (h *NotificationHandler) GetTemplate(c *gin.Context) {
	id := c.Param("id")

	template, err := h.domain.Notification().GetTemplate(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, template)
}

func (h *NotificationHandler) ListTemplates(c *gin.Context) {
	templates, err := h.domain.Notification().ListTemplates(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, templates)
}

func (h *NotificationHandler) UpdateTemplate(c *gin.Context) {
	id := c.Param("id")
	var input model.NotificationTemplateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	template, err := h.domain.Notification().UpdateTemplate(c.Request.Context(), id, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, template)
}

func (h *NotificationHandler) DeleteTemplate(c *gin.Context) {
	id := c.Param("id")

	err := h.domain.Notification().DeleteTemplate(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "template deleted successfully"})
}

func (h *NotificationHandler) SendPaymentConfirmationNotification(c *gin.Context) {
	type Request struct {
		CustomerID string  `json:"customer_id" binding:"required"`
		PaymentID  string  `json:"payment_id" binding:"required"`
		Amount     float64 `json:"amount" binding:"required"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.domain.Notification().SendPaymentConfirmation(c.Request.Context(), req.CustomerID, req.PaymentID, req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "payment confirmation sent successfully"})
}

func (h *NotificationHandler) SendInvoiceReminderNotification(c *gin.Context) {
	type Request struct {
		CustomerID string `json:"customer_id" binding:"required"`
		InvoiceID  string `json:"invoice_id" binding:"required"`
		Days       int    `json:"days" binding:"required,min=1"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.domain.Notification().SendInvoiceReminder(c.Request.Context(), req.CustomerID, req.InvoiceID, req.Days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "invoice reminder sent successfully"})
}

func (h *NotificationHandler) SendPaymentFailedNotification(c *gin.Context) {
	type Request struct {
		CustomerID string `json:"customer_id" binding:"required"`
		InvoiceID  string `json:"invoice_id" binding:"required"`
		Reason     string `json:"reason" binding:"required"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.domain.Notification().SendPaymentFailed(c.Request.Context(), req.CustomerID, req.InvoiceID, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "payment failed notification sent successfully"})
}

func (h *NotificationHandler) SendInvoiceCreatedNotification(c *gin.Context) {
	type Request struct {
		CustomerID string `json:"customer_id" binding:"required"`
		InvoiceID  string `json:"invoice_id" binding:"required"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.domain.Notification().SendInvoiceCreated(c.Request.Context(), req.CustomerID, req.InvoiceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "invoice created notification sent successfully"})
}
