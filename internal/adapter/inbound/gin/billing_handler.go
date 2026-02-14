package gin_inbound_adapter

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go-template/internal/domain"
	"go-template/internal/model"
	inbound_port "go-template/internal/port/inbound"
)

type BillingHandler struct {
	domain domain.Domain
}

func NewBillingHandler(domain domain.Domain) inbound_port.BillingHttpPort {
	return &BillingHandler{domain: domain}
}

func (h *BillingHandler) CreateInvoice(c *gin.Context) {
	var input model.InvoiceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	invoice, err := h.domain.Billing().CreateInvoice(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, invoice)
}

func (h *BillingHandler) GetInvoice(c *gin.Context) {
	id := c.Param("id")

	invoice, err := h.domain.Billing().GetInvoice(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, invoice)
}

func (h *BillingHandler) ListInvoices(c *gin.Context) {
	var filter model.InvoiceFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	invoices, err := h.domain.Billing().ListInvoices(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, invoices)
}

func (h *BillingHandler) UpdateInvoice(c *gin.Context) {
	id := c.Param("id")
	var input model.InvoiceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	invoice, err := h.domain.Billing().UpdateInvoice(c.Request.Context(), id, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, invoice)
}

func (h *BillingHandler) DeleteInvoice(c *gin.Context) {
	id := c.Param("id")

	err := h.domain.Billing().DeleteInvoice(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "invoice deleted successfully"})
}

func (h *BillingHandler) GenerateMonthlyInvoices(c *gin.Context) {
	type GenerateRequest struct {
		Year  int `json:"year" binding:"required"`
		Month int `json:"month" binding:"required,min=1,max=12"`
	}

	var req GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	results, err := h.domain.Billing().GenerateMonthlyInvoices(c.Request.Context(), req.Year, req.Month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count":   len(results),
		"results": results,
	})
}

func (h *BillingHandler) ApplyPayment(c *gin.Context) {
	id := c.Param("id")

	type PaymentRequest struct {
		Amount float64 `json:"amount" binding:"required,min=0"`
	}

	var req PaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.domain.Billing().ApplyPayment(c.Request.Context(), id, req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "payment applied successfully"})
}

func (h *BillingHandler) CheckOverdueInvoices(c *gin.Context) {
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

func (h *BillingHandler) CalculateLateFee(c *gin.Context) {
	id := c.Param("id")

	invoice, err := h.domain.Billing().GetInvoice(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	lateFee := h.domain.Billing().CalculateLateFee(c.Request.Context(), invoice)

	c.JSON(http.StatusOK, gin.H{
		"invoice_id": invoice.ID,
		"late_fee":   lateFee,
	})
}

func (h *BillingHandler) CreateInvoiceItem(c *gin.Context) {
	var input model.InvoiceItemInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item, err := h.domain.Billing().CreateInvoiceItem(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, item)
}

func (h *BillingHandler) GetInvoiceItem(c *gin.Context) {
	id := c.Param("id")

	item, err := h.domain.Billing().GetInvoiceItem(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *BillingHandler) ListInvoiceItems(c *gin.Context) {
	invoiceID := c.Query("invoice_id")
	if invoiceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invoice_id is required"})
		return
	}

	items, err := h.domain.Billing().ListInvoiceItems(c.Request.Context(), invoiceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, items)
}
