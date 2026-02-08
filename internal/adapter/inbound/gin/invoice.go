package gin_inbound_adapter

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/palantir/stacktrace"

	"mikrops/internal/domain"
	"mikrops/internal/model"
	inbound_port "mikrops/internal/port/inbound"
	"mikrops/utils/activity"
)

type invoiceAdapter struct {
	domain domain.Domain
}

func NewInvoiceAdapter(domain domain.Domain) inbound_port.InvoiceHttpPort {
	return &invoiceAdapter{domain: domain}
}

func (h *invoiceAdapter) List(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_invoice_list")

	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(string)

	filter := model.InvoiceFilter{
		TenantIDs:        []string{tid},
		WithCustomer:     true,
		WithSubscription: true,
	}

	results, err := h.domain.Invoice().FindByFilter(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: results})
	return nil
}

func (h *invoiceAdapter) Create(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_invoice_create")

	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(string)

	var payload model.InvoiceInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return nil
	}
	payload.TenantID = tid

	result, err := h.domain.Invoice().Create(ctx, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusCreated, model.Response{Success: true, Data: result})
	return nil
}

func (h *invoiceAdapter) Get(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_invoice_get")

	id := c.Param("id")

	result, err := h.domain.Invoice().FindByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: result})
	return nil
}

func (h *invoiceAdapter) Update(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_invoice_update")

	id := c.Param("id")

	var payload model.InvoiceInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return nil
	}

	err := h.domain.Invoice().Update(ctx, id, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true})
	return nil
}

func (h *invoiceAdapter) GenerateBulk(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_invoice_generate_bulk")

	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(string)

	var payload struct {
		PeriodStart time.Time `json:"period_start" binding:"required"`
		PeriodEnd   time.Time `json:"period_end" binding:"required"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return nil
	}

	err := h.domain.Invoice().GenerateBulk(ctx, tid, payload.PeriodStart, payload.PeriodEnd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusCreated, model.Response{Success: true})
	return nil
}
