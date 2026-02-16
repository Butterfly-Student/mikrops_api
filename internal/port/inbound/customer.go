package inbound_port

import "github.com/gin-gonic/gin"

type CustomerHttpPort interface {
	CreateCustomer(c *gin.Context)
	GetCustomer(c *gin.Context)
	GetCustomerByCode(c *gin.Context)
	ListCustomers(c *gin.Context)
	UpdateCustomer(c *gin.Context)
	DeleteCustomer(c *gin.Context)
	GetBillingInfo(c *gin.Context)
	IsolateCustomer(c *gin.Context)
	ActivateCustomer(c *gin.Context)
}
