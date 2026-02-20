package gin_inbound_adapter

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/palantir/stacktrace"

	"go-template/internal/domain"
	"go-template/internal/model"
	inbound_port "go-template/internal/port/inbound"
	"go-template/utils/activity"
)

type invoiceAdapter struct {
	domain domain.Domain
}

func NewInvoiceAdapter(
	domain domain.Domain,
) inbound_port.InvoiceHttpPort {
	return &invoiceAdapter{
		domain: domain,
	}
}

func (h *invoiceAdapter) Create(c *gin.Context) {
	ctx := activity.NewContext("http_invoice_create")

	var input model.InvoiceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	ctx = activity.WithPayload(ctx, input)

	invoice, err := h.domain.Invoice().Create(ctx, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, model.Response{
		Success: true,
		Data:    invoice,
	})
}

func (h *invoiceAdapter) GetByID(c *gin.Context) {
	ctx := activity.NewContext("http_invoice_get_by_id")
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   "id parameter is required",
		})
		return
	}

	ctx = activity.WithPayload(ctx, map[string]string{"id": id})

	invoice, err := h.domain.Invoice().GetByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Success: true,
		Data:    invoice,
	})
}

func (h *invoiceAdapter) GetByNumber(c *gin.Context) {
	ctx := activity.NewContext("http_invoice_get_by_number")
	number := c.Param("number")

	if number == "" {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   "number parameter is required",
		})
		return
	}

	ctx = activity.WithPayload(ctx, map[string]string{"number": number})

	invoice, err := h.domain.Invoice().GetByNumber(ctx, number)
	if err != nil {
		c.JSON(http.StatusNotFound, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Success: true,
		Data:    invoice,
	})
}

func (h *invoiceAdapter) List(c *gin.Context) {
	ctx := activity.NewContext("http_invoice_list")

	var filter model.InvoiceFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	ctx = activity.WithPayload(ctx, filter)

	invoices, err := h.domain.Invoice().List(ctx, &filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Success: true,
		Data:    invoices,
	})
}

func (h *invoiceAdapter) Update(c *gin.Context) {
	ctx := activity.NewContext("http_invoice_update")
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   "id parameter is required",
		})
		return
	}

	var input model.InvoiceInput
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

	invoice, err := h.domain.Invoice().Update(ctx, id, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Success: true,
		Data:    invoice,
	})
}

func (h *invoiceAdapter) Delete(c *gin.Context) {
	ctx := activity.NewContext("http_invoice_delete")
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   "id parameter is required",
		})
		return
	}

	ctx = activity.WithPayload(ctx, map[string]string{"id": id})

	if err := h.domain.Invoice().Delete(ctx, id); err != nil {
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

func (h *invoiceAdapter) GenerateMonthly(c *gin.Context) {
	ctx := activity.NewContext("http_invoice_generate_monthly")
	customerID := c.Param("customer_id")

	if customerID == "" {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   "customer_id parameter is required",
		})
		return
	}

	// Get month and year from query params (default to current month/year)
	monthStr := c.DefaultQuery("month", "")
	yearStr := c.DefaultQuery("year", "")

	var month, year int
	var err error

	if monthStr != "" {
		month, err = strconv.Atoi(monthStr)
		if err != nil || month < 1 || month > 12 {
			c.JSON(http.StatusBadRequest, model.Response{
				Success: false,
				Error:   "invalid month parameter",
			})
			return
		}
	} else {
		month = int(time.Now().Month())
	}

	if yearStr != "" {
		year, err = strconv.Atoi(yearStr)
		if err != nil || year < 2000 {
			c.JSON(http.StatusBadRequest, model.Response{
				Success: false,
				Error:   "invalid year parameter",
			})
			return
		}
	} else {
		year = time.Now().Year()
	}

	ctx = activity.WithPayload(ctx, map[string]interface{}{
		"customer_id": customerID,
		"month":       month,
		"year":        year,
	})

	invoice, err := h.domain.Invoice().GenerateMonthlyInvoice(ctx, customerID, month, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, model.Response{
		Success: true,
		Data:    invoice,
	})
}

func (h *invoiceAdapter) CalculateLateFee(c *gin.Context) {
	ctx := activity.NewContext("http_invoice_calculate_late_fee")
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, model.Response{
			Success: false,
			Error:   "id parameter is required",
		})
		return
	}

	ctx = activity.WithPayload(ctx, map[string]string{"id": id})

	invoice, err := h.domain.Invoice().CalculateLateFee(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Success: false,
			Error:   stacktrace.RootCause(err).Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Success: true,
		Data:    invoice,
	})
}
