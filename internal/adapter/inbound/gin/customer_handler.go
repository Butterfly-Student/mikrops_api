package gin_inbound_adapter

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go-template/internal/domain"
	"go-template/internal/model"
	inbound_port "go-template/internal/port/inbound"
)

type CustomerHandler struct {
	domain domain.Domain
}

func NewCustomerHandler(domain domain.Domain) inbound_port.CustomerHttpPort {
	return &CustomerHandler{domain: domain}
}

func (h *CustomerHandler) CreateCustomer(c *gin.Context) {
	var input model.CustomerInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	customer, err := h.domain.Customer().CreateCustomer(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, customer)
}

func (h *CustomerHandler) GetCustomer(c *gin.Context) {
	id := c.Param("id")

	customer, err := h.domain.Customer().GetCustomer(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, customer)
}

func (h *CustomerHandler) ListCustomers(c *gin.Context) {
	var filter model.CustomerFilter

	if statuses := c.QueryArray("status"); len(statuses) > 0 {
		filter.Status = statuses
	}
	if search := c.Query("search"); search != "" {
		filter.Search = &search
	}
	if isExpired := c.Query("is_expired"); isExpired != "" {
		isExpiredBool := isExpired == "true"
		filter.IsExpired = &isExpiredBool
	}

	customers, err := h.domain.Customer().ListCustomers(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, customers)
}

func (h *CustomerHandler) UpdateCustomer(c *gin.Context) {
	id := c.Param("id")

	var input model.CustomerInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	customer, err := h.domain.Customer().UpdateCustomer(c.Request.Context(), id, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, customer)
}

func (h *CustomerHandler) DeleteCustomer(c *gin.Context) {
	id := c.Param("id")

	err := h.domain.Customer().DeleteCustomer(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Customer deleted successfully"})
}

func (h *CustomerHandler) GetBillingInfo(c *gin.Context) {
	id := c.Param("id")

	customer, err := h.domain.Customer().GetCustomer(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	billingInfo := gin.H{
		"customer_id":       customer.ID,
		"customer_code":     customer.CustomerCode,
		"full_name":         customer.FullName,
		"expiry_date":       customer.ExpiryDate,
		"billing_cycle":     customer.BillingCycle,
		"billing_day":       customer.BillingDay,
		"auto_isolate":      customer.AutoIsolate,
		"grace_period_days": customer.GracePeriodDays,
		"profile":           customer.Profile,
		"router":            customer.Router,
	}

	c.JSON(http.StatusOK, billingInfo)
}

func (h *CustomerHandler) IsolateCustomer(c *gin.Context) {
	id := c.Param("id")

	customer, err := h.domain.Customer().GetCustomer(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if customer.Status == "isolated" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Customer is already isolated"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Customer isolation initiated"})
}

func (h *CustomerHandler) ActivateCustomer(c *gin.Context) {
	id := c.Param("id")

	customer, err := h.domain.Customer().GetCustomer(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if customer.Status == "active" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Customer is already active"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Customer activation initiated"})
}
