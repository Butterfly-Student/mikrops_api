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

type paymentMethodAdapter struct {
	domain domain.Domain
}

func NewPaymentMethodAdapter(domain domain.Domain) inbound_port.PaymentMethodHttpPort {
	return &paymentMethodAdapter{domain: domain}
}

func (h *paymentMethodAdapter) List(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_payment_method_list")

	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(string)

	filter := model.PaymentMethodFilter{TenantIDs: []string{tid}}

	results, err := h.domain.PaymentMethod().FindByFilter(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: results})
	return nil
}

func (h *paymentMethodAdapter) Create(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_payment_method_create")

	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(string)

	var payload model.PaymentMethodInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return nil
	}
	payload.TenantID = tid

	result, err := h.domain.PaymentMethod().Create(ctx, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusCreated, model.Response{Success: true, Data: result})
	return nil
}

func (h *paymentMethodAdapter) Get(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_payment_method_get")

	id := c.Param("id")

	result, err := h.domain.PaymentMethod().FindByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: result})
	return nil
}

func (h *paymentMethodAdapter) Update(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_payment_method_update")

	id := c.Param("id")

	var payload model.PaymentMethodInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return nil
	}

	err := h.domain.PaymentMethod().Update(ctx, id, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true})
	return nil
}

func (h *paymentMethodAdapter) Delete(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_payment_method_delete")

	id := c.Param("id")

	err := h.domain.PaymentMethod().Delete(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true})
	return nil
}
