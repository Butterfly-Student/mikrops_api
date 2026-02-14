package inbound_port

import "github.com/gin-gonic/gin"

type PaymentHttpPort interface {
	CreatePayment(c *gin.Context)
	GetPayment(c *gin.Context)
	ListPayments(c *gin.Context)
	UpdatePayment(c *gin.Context)
	DeletePayment(c *gin.Context)
	ProcessXenditWebhook(c *gin.Context)
	CreatePaymentLink(c *gin.Context)
	GenerateReceipt(c *gin.Context)
	GetPaymentHistory(c *gin.Context)
	GetPaymentStatistics(c *gin.Context)
}
