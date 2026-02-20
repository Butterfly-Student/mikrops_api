package inbound_port

import "github.com/gin-gonic/gin"

type InvoiceHttpPort interface {
	Create(c *gin.Context)
	GetByID(c *gin.Context)
	GetByNumber(c *gin.Context)
	List(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
	GenerateMonthly(c *gin.Context)
	CalculateLateFee(c *gin.Context)
}
