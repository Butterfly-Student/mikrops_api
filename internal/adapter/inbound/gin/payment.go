package gin_inbound_adapter

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/palantir/stacktrace"

	"go-template/internal/domain"
	"go-template/internal/model"
	inbound_port "go-template/internal/port/inbound"
	"go-template/utils/activity"
)

type paymentAdapter struct {
	domain domain.Domain
}

func NewPaymentAdapter(
	domain domain.Domain,
) inbound_port.PaymentHttpPort {
	return &paymentAdapter{
		domain: domain,
	}
}

func (h *paymentAdapter) Create(c *gin.Context) {
	ctx := activity.NewContext("http_payment_create")

	var input model.PaymentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	ctx = activity.WithPayload(ctx, input)

	payment, err := h.domain.Payment().Create(ctx, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, model.Response{
		Success: true,
		Data:    payment,
	})
}

func (h *paymentAdapter) GetByID(c *gin.Context) {
	ctx := activity.NewContext("http_payment_get_by_id")
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   "id parameter is required",
		})
		return
	}

	ctx = activity.WithPayload(ctx, map[string]string{"id": id})

	payment, err := h.domain.Payment().GetByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Success: true,
		Data:    payment,
	})
}

func (h *paymentAdapter) GetByNumber(c *gin.Context) {
	ctx := activity.NewContext("http_payment_get_by_number")
	number := c.Param("number")

	if number == "" {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   "number parameter is required",
		})
		return
	}

	ctx = activity.WithPayload(ctx, map[string]string{"number": number})

	payment, err := h.domain.Payment().GetByNumber(ctx, number)
	if err != nil {
		c.JSON(http.StatusNotFound, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Success: true,
		Data:    payment,
	})
}

func (h *paymentAdapter) List(c *gin.Context) {
	ctx := activity.NewContext("http_payment_list")

	var filter model.PaymentFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	ctx = activity.WithPayload(ctx, filter)

	payments, err := h.domain.Payment().List(ctx, &filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Success: true,
		Data:    payments,
	})
}

func (h *paymentAdapter) Update(c *gin.Context) {
	ctx := activity.NewContext("http_payment_update")
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   "id parameter is required",
		})
		return
	}

	var input model.PaymentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	ctx = activity.WithPayload(ctx, map[string]interface{}{
		"id":    id,
		"input": input,
	})

	payment, err := h.domain.Payment().Update(ctx, id, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Success: true,
		Data:    payment,
	})
}

func (h *paymentAdapter) Delete(c *gin.Context) {
	ctx := activity.NewContext("http_payment_delete")
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   "id parameter is required",
		})
		return
	}

	ctx = activity.WithPayload(ctx, map[string]string{"id": id})

	if err := h.domain.Payment().Delete(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Success: true,
		Data:    nil,
	})
}

func (h *paymentAdapter) Confirm(c *gin.Context) {
	ctx := activity.NewContext("http_payment_confirm")
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   "id parameter is required",
		})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, model.Response{
			Success: false,
			Error:   "user not authenticated",
		})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, model.Response{
			Success: false,
			Error:   "invalid user id format",
		})
		return
	}

	ctx = activity.WithPayload(ctx, map[string]string{
		"id":      id,
		"user_id": userIDStr,
	})

	payment, err := h.domain.Payment().Confirm(ctx, id, userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Success: true,
		Data:    payment,
	})
}

func (h *paymentAdapter) Reject(c *gin.Context) {
	ctx := activity.NewContext("http_payment_reject")
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   "id parameter is required",
		})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, model.Response{
			Success: false,
			Error:   "user not authenticated",
		})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, model.Response{
			Success: false,
			Error:   "invalid user id format",
		})
		return
	}

	// Get rejection reason from request body
	var requestBody struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	ctx = activity.WithPayload(ctx, map[string]string{
		"id":      id,
		"user_id": userIDStr,
		"reason":  requestBody.Reason,
	})

	payment, err := h.domain.Payment().Reject(ctx, id, userIDStr, requestBody.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Success: true,
		Data:    payment,
	})
}

func (h *paymentAdapter) Allocate(c *gin.Context) {
	ctx := activity.NewContext("http_payment_allocate")
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   "id parameter is required",
		})
		return
	}

	// Get allocation details from request body
	var requestBody struct {
		InvoiceID string  `json:"invoice_id" binding:"required"`
		Amount    float64 `json:"amount" binding:"required,min=0"`
	}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	ctx = activity.WithPayload(ctx, map[string]interface{}{
		"payment_id": id,
		"invoice_id": requestBody.InvoiceID,
		"amount":     requestBody.Amount,
	})

	allocation, err := h.domain.Payment().AllocateToInvoice(ctx, id, requestBody.InvoiceID, requestBody.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, model.Response{
		Success: true,
		Data:    allocation,
	})
}
