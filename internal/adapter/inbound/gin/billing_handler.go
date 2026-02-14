package gin_inbound_adapter

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"go-template/internal/domain"
	"go-template/internal/model"
	inbound_port "go-template/internal/port/inbound"
)

type billingAdapter struct {
	domain domain.Domain
}

func NewBillingAdapter(domain domain.Domain) inbound_port.BillingHttpPort {
	return &billingAdapter{
		domain: domain,
	}
}

func (h *billingAdapter) CreateInvoice(c *gin.Context) {
	var req model.InvoiceInput
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	invoice, err := h.domain.Billing().CreateInvoice(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, invoice)
}

func (h *billingAdapter) GetInvoice(c *gin.Context) {
	id := c.Param("id")

	invoice, err := h.domain.Billing().FindByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, invoice)
}

func (h *billingAdapter) ListInvoices(c *gin.Context) {
	var filter model.InvoiceFilter

	// Parse query parameters
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	invoices, err := h.domain.Billing().FindByFilter(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, invoices)
}

func (h *billingAdapter) UpdateInvoiceStatus(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Status model.InvoiceStatus `json:"status" binding:"required,oneof=draft sent partial paid overdue cancelled refunded"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.domain.Billing().UpdateStatus(c.Request.Context(), id, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Invoice status updated successfully"})
}

func (h *billingAdapter) CancelInvoice(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.domain.Billing().CancelInvoice(c.Request.Context(), id, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Invoice cancelled successfully"})
}

func (h *billingAdapter) GenerateMonthlyInvoices(c *gin.Context) {
	var req struct {
		Year  int `json:"year" binding:"required,min=2000,max=2100"`
		Month int `json:"month" binding:"required,min=1,max=12"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	invoices, err := h.domain.Billing().GenerateMonthlyInvoices(c.Request.Context(), req.Year, req.Month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Monthly invoices generated successfully",
		"count":   len(invoices),
		"invoices": invoices,
	})
}

func (h *billingAdapter) ApplyPayment(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Amount        string    `json:"amount" binding:"required"`
		PaymentDate   time.Time `json:"payment_date" binding:"required"`
		PaymentMethod string    `json:"payment_method" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse amount
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid amount format"})
		return
	}

	err = h.domain.Billing().ApplyPayment(c.Request.Context(), id, amount, req.PaymentDate, req.PaymentMethod)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Payment applied successfully"})
}

func (h *billingAdapter) GetOverdueInvoices(c *gin.Context) {
	invoices, err := h.domain.Billing().CheckOverdueInvoices(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count":    len(invoices),
		"invoices": invoices,
	})
}

func (h *billingAdapter) GetMonthlyRevenue(c *gin.Context) {
	// Get year and month from query params
	yearStr := c.Query("year")
	monthStr := c.Query("month")

	// Default to current year and month if not provided
	now := time.Now()
	year := now.Year()
	month := int(now.Month())

	if yearStr != "" {
		parsedYear, err := strconv.Atoi(yearStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year format"})
			return
		}
		year = parsedYear
	}

	if monthStr != "" {
		parsedMonth, err := strconv.Atoi(monthStr)
		if err != nil || parsedMonth < 1 || parsedMonth > 12 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid month format"})
			return
		}
		month = parsedMonth
	}

	revenue, err := h.domain.Billing().GetMonthlyRevenue(c.Request.Context(), year, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"year":    year,
		"month":   month,
		"revenue": revenue,
	})
}
