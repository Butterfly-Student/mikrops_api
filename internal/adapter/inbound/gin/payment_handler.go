package gin_inbound_adapter

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"go-template/internal/domain"
	"go-template/internal/model"
	inbound_port "go-template/internal/port/inbound"
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
		ExternalID string `json:"external_id"`
		Status     string `json:"status"`
	}

	if err := c.ShouldBindJSON(&webhookData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "webhook received"})
}

func (h *PaymentHandler) CreatePaymentLink(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "payment link creation not implemented yet"})
}
