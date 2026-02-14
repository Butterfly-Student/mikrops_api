package inbound_port

import "github.com/gin-gonic/gin"

type BillingHttpPort interface {
	CreateInvoice(c *gin.Context)
	GetInvoice(c *gin.Context)
	ListInvoices(c *gin.Context)
	UpdateInvoice(c *gin.Context)
	DeleteInvoice(c *gin.Context)
	GenerateMonthlyInvoices(c *gin.Context)
	ApplyPayment(c *gin.Context)
	CheckOverdueInvoices(c *gin.Context)
	CalculateLateFee(c *gin.Context)
	CreateInvoiceItem(c *gin.Context)
	GetInvoiceItem(c *gin.Context)
	ListInvoiceItems(c *gin.Context)
}
