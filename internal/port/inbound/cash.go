package inbound_port

import "github.com/gin-gonic/gin"

type CashHttpPort interface {
	CreateCashCategory(c *gin.Context)
	GetCashCategory(c *gin.Context)
	ListCashCategories(c *gin.Context)
	UpdateCashCategory(c *gin.Context)
	DeleteCashCategory(c *gin.Context)
	CreateCashTransaction(c *gin.Context)
	GetCashTransaction(c *gin.Context)
	ListCashTransactions(c *gin.Context)
	UpdateCashTransaction(c *gin.Context)
	DeleteCashTransaction(c *gin.Context)
	ApproveCashTransaction(c *gin.Context)
	RejectCashTransaction(c *gin.Context)
	GetCashBalance(c *gin.Context)
}
