package gin_inbound_adapter

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"go-template/internal/domain"
	"go-template/internal/model"
	inbound_port "go-template/internal/port/inbound"
)

type customerAdapter struct {
	domain domain.Domain
}

func NewCustomerAdapter(domain domain.Domain) inbound_port.CustomerHttpPort {
	return &customerAdapter{
		domain: domain,
	}
}

func (h *customerAdapter) Create(c *gin.Context) {
	var req model.CustomerInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	customer, err := h.domain.Customer().Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, customer)
}

func (h *customerAdapter) Get(c *gin.Context) {
	id := c.Param("id")

	customer, err := h.domain.Customer().FindByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, customer)
}

func (h *customerAdapter) List(c *gin.Context) {
	var filter model.CustomerFilter

	// Parse query parameters
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	customers, err := h.domain.Customer().FindByFilter(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, customers)
}

func (h *customerAdapter) Update(c *gin.Context) {
	id := c.Param("id")

	var req model.CustomerInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	customer, err := h.domain.Customer().Update(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, customer)
}

func (h *customerAdapter) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.domain.Customer().Delete(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Customer deleted successfully"})
}

func (h *customerAdapter) GetBillingInfo(c *gin.Context) {
	id := c.Param("id")

	billingInfo, err := h.domain.Customer().GetBillingInfo(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, billingInfo)
}

func (h *customerAdapter) Isolate(c *gin.Context) {
	id := c.Param("id")

	err := h.domain.Customer().Isolate(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Customer isolated successfully"})
}

func (h *customerAdapter) Activate(c *gin.Context) {
	id := c.Param("id")

	err := h.domain.Customer().Activate(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Customer activated successfully"})
}
