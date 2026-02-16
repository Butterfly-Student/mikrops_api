package inbound_port

import "github.com/gin-gonic/gin"

type RefundHttpPort interface {
	CreateRefund(c *gin.Context)
	GetRefund(c *gin.Context)
	ListRefunds(c *gin.Context)
	UpdateRefund(c *gin.Context)
	DeleteRefund(c *gin.Context)
	ApproveRefund(c *gin.Context)
	RejectRefund(c *gin.Context)
	ProcessRefund(c *gin.Context)
	CompleteRefund(c *gin.Context)
	GetPendingRefunds(c *gin.Context)
}
