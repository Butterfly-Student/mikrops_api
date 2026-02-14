package inbound_port

import "github.com/gin-gonic/gin"

type NotificationHttpPort interface {
	CreateNotification(c *gin.Context)
	GetNotification(c *gin.Context)
	ListNotifications(c *gin.Context)
	SendNotification(c *gin.Context)
	RetryFailedNotifications(c *gin.Context)

	// Notification Templates
	CreateTemplate(c *gin.Context)
	GetTemplate(c *gin.Context)
	ListTemplates(c *gin.Context)
	UpdateTemplate(c *gin.Context)
	DeleteTemplate(c *gin.Context)

	// Predefined Notification Types
	SendPaymentConfirmationNotification(c *gin.Context)
	SendInvoiceReminderNotification(c *gin.Context)
	SendPaymentFailedNotification(c *gin.Context)
	SendInvoiceCreatedNotification(c *gin.Context)
}
