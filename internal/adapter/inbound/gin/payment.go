package gin_inbound_adapter

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/palantir/stacktrace"

	"mikrops/internal/domain"
	"mikrops/internal/model"
	inbound_port "mikrops/internal/port/inbound"
	"mikrops/utils/activity"
)

type paymentAdapter struct {
	domain domain.Domain
}

func NewPaymentAdapter(domain domain.Domain) inbound_port.PaymentHttpPort {
	return &paymentAdapter{domain: domain}
}

func (h *paymentAdapter) List(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_payment_list")

	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(string)

	filter := model.PaymentFilter{
		TenantIDs:         []string{tid},
		WithInvoice:       true,
		WithPaymentMethod: true,
	}

	results, err := h.domain.Payment().FindByFilter(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: results})
	return nil
}

func (h *paymentAdapter) Get(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_payment_get")

	id := c.Param("id")

	result, err := h.domain.Payment().FindByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: result})
	return nil
}

func (h *paymentAdapter) Verify(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_payment_verify")

	id := c.Param("id")
	staffID, _ := c.Get("staff_id")
	sid, _ := staffID.(string)

	err := h.domain.Payment().Verify(ctx, id, sid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true})
	return nil
}

func (h *paymentAdapter) Reject(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_payment_reject")

	id := c.Param("id")
	staffID, _ := c.Get("staff_id")
	sid, _ := staffID.(string)

	var payload struct {
		Notes string `json:"notes"`
	}
	_ = c.ShouldBindJSON(&payload)

	err := h.domain.Payment().Reject(ctx, id, sid, payload.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true})
	return nil
}
