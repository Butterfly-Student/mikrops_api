package inbound_port

import "github.com/gin-gonic/gin"

type BillingHttpPort interface {
	CreateInvoice(c *gin.Context)
	GetInvoice(c *gin.Context)
	ListInvoices(c *gin.Context)
	UpdateInvoiceStatus(c *gin.Context)
	CancelInvoice(c *gin.Context)
	GenerateMonthlyInvoices(c *gin.Context)
	ApplyPayment(c *gin.Context)
	GetOverdueInvoices(c *gin.Context)
	GetMonthlyRevenue(c *gin.Context)
}
