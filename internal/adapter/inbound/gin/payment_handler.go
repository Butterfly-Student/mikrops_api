package gin_inbound_adapter

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"go-template/internal/domain"
	"go-template/internal/model"
	inbound_port "go-template/internal/port/inbound"
	"go-template/utils/log"
	"go-template/utils/xendit"
)

type PaymentHandler struct {
	domain domain.Domain
}

func NewPaymentAdapter(
	domain domain.Domain,
) inbound_port.PaymentHttpPort {
	return &PaymentHandler{
		domain: domain,
	}
}

func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	var input model.PaymentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payment, err := h.domain.Payment().CreatePayment(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, payment)
}

func (h *PaymentHandler) GetPayment(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payment ID is required"})
		return
	}

	payment, err := h.domain.Payment().GetPayment(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, payment)
}

func (h *PaymentHandler) ListPayments(c *gin.Context) {
	filter := model.PaymentFilter{}

	if customerIDs := c.QueryArray("customer_id"); len(customerIDs) > 0 {
		ids := make([]string, len(customerIDs))
		copy(ids, customerIDs)
	}

	if status := c.Query("status"); status != "" {
		filter.Status = append(filter.Status, status)
	}

	if paymentMethod := c.Query("payment_method"); paymentMethod != "" {
		filter.PaymentMethod = append(filter.PaymentMethod, paymentMethod)
	}

	if search := c.Query("search"); search != "" {
		filter.Search = &search
	}

	if startDate := c.Query("start_date"); startDate != "" {
		t, err := time.Parse(time.RFC3339, startDate)
		if err == nil {
			filter.PaymentStart = &t
		}
	}

	if endDate := c.Query("end_date"); endDate != "" {
		t, err := time.Parse(time.RFC3339, endDate)
		if err == nil {
			filter.PaymentEnd = &t
		}
	}

	if isProcessed := c.Query("is_processed"); isProcessed != "" {
		processed, err := strconv.ParseBool(isProcessed)
		if err == nil {
			filter.IsProcessed = &processed
		}
	}

	payments, err := h.domain.Payment().ListPayments(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, payments)
}

func (h *PaymentHandler) UpdatePayment(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payment ID is required"})
		return
	}

	var input model.PaymentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payment, err := h.domain.Payment().UpdatePayment(c.Request.Context(), id, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, payment)
}

func (h *PaymentHandler) DeletePayment(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payment ID is required"})
		return
	}

	err := h.domain.Payment().DeletePayment(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "payment deleted successfully"})
}

func (h *PaymentHandler) ProcessXenditWebhook(c *gin.Context) {
	var webhookData struct {
		ExternalID string  `json:"external_id"`
		Status     string  `json:"status"`
		PaymentID  string  `json:"payment_id"`
		PaidAmount float64 `json:"paid_amount"`
	}

	if err := c.ShouldBindJSON(&webhookData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Process webhook using domain
	err := h.domain.Payment().ProcessWebhook(c.Request.Context(), &xendit.WebhookData{
		ExternalID: webhookData.ExternalID,
		Status:     webhookData.Status,
		PaymentID:  webhookData.PaymentID,
		PaidAmount: webhookData.PaidAmount,
	})
	if err != nil {
		log.WithContext(c.Request.Context()).Error(fmt.Sprintf("Failed to process webhook: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "webhook processed successfully"})
}

func (h *PaymentHandler) CreatePaymentLink(c *gin.Context) {
	var input struct {
		InvoiceID     string `json:"invoice_id" binding:"required"`
		PaymentMethod string `json:"payment_method"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	invoice, err := h.domain.Billing().GetInvoice(c.Request.Context(), input.InvoiceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "invoice not found"})
		return
	}

	customer, err := h.domain.Customer().GetCustomer(c.Request.Context(), invoice.CustomerID.String())
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "customer not found"})
		return
	}

	xenditInvoice, err := h.domain.Payment().CreateXenditInvoice(c.Request.Context(), invoice, customer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"payment_url": xenditInvoice.InvoiceURL,
		"external_id": xenditInvoice.ExternalID,
		"amount":      xenditInvoice.Amount,
	})
}

// GenerateReceipt generates a PDF receipt for a payment
func (h *PaymentHandler) GenerateReceipt(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payment ID is required"})
		return
	}

	pdfBytes, err := h.domain.Payment().GenerateReceipt(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get payment to use number in filename
	payment, err := h.domain.Payment().GetPayment(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	filename := "receipt_" + payment.PaymentNumber + ".pdf"

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Length", strconv.Itoa(len(pdfBytes)))

	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

// GetPaymentHistory returns payment history for a customer
func (h *PaymentHandler) GetPaymentHistory(c *gin.Context) {
	customerID := c.Param("customer_id")
	if customerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "customer ID is required"})
		return
	}

	filter := model.PaymentFilter{}

	if status := c.Query("status"); status != "" {
		filter.Status = append(filter.Status, status)
	}

	if paymentMethod := c.Query("payment_method"); paymentMethod != "" {
		filter.PaymentMethod = append(filter.PaymentMethod, paymentMethod)
	}

	if startDate := c.Query("start_date"); startDate != "" {
		t, err := time.Parse(time.RFC3339, startDate)
		if err == nil {
			filter.PaymentStart = &t
		}
	}

	if endDate := c.Query("end_date"); endDate != "" {
		t, err := time.Parse(time.RFC3339, endDate)
		if err == nil {
			filter.PaymentEnd = &t
		}
	}

	payments, totalCount, err := h.domain.Payment().GetPaymentHistory(c.Request.Context(), customerID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"payments":    payments,
		"total_count": totalCount,
	})
}

// GetPaymentStatistics returns payment statistics for a customer
func (h *PaymentHandler) GetPaymentStatistics(c *gin.Context) {
	customerID := c.Param("customer_id")
	if customerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "customer ID is required"})
		return
	}

	stats, err := h.domain.Payment().GetPaymentStatistics(c.Request.Context(), customerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}
