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

type pppoeAccountAdapter struct {
	domain domain.Domain
}

func NewPppoeAccountAdapter(domain domain.Domain) inbound_port.PppoeAccountHttpPort {
	return &pppoeAccountAdapter{domain: domain}
}

func (h *pppoeAccountAdapter) List(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_pppoe_account_list")

	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(string)

	filter := model.PppoeAccountFilter{
		TenantIDs:    []string{tid},
		WithCustomer: true,
		WithNas:      true,
		WithPackage:  true,
	}

	results, err := h.domain.PppoeAccount().FindByFilter(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: results})
	return nil
}

func (h *pppoeAccountAdapter) Get(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_pppoe_account_get")

	id := c.Param("id")

	result, err := h.domain.PppoeAccount().FindByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: result})
	return nil
}

func (h *pppoeAccountAdapter) Create(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_pppoe_account_create")

	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(string)

	var payload model.PppoeAccountInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return nil
	}
	payload.TenantID = tid

	result, err := h.domain.PppoeAccount().Create(ctx, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusCreated, model.Response{Success: true, Data: result})
	return nil
}

func (h *pppoeAccountAdapter) Update(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_pppoe_account_update")

	id := c.Param("id")

	var payload model.PppoeAccountInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return nil
	}

	err := h.domain.PppoeAccount().Update(ctx, id, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true})
	return nil
}

func (h *pppoeAccountAdapter) Delete(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_pppoe_account_delete")

	id := c.Param("id")

	err := h.domain.PppoeAccount().Delete(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true})
	return nil
}

func (h *pppoeAccountAdapter) Isolate(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_pppoe_account_isolate")

	id := c.Param("id")

	result, err := h.domain.PppoeAccount().Isolate(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: result})
	return nil
}

func (h *pppoeAccountAdapter) Restore(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_pppoe_account_restore")

	id := c.Param("id")

	result, err := h.domain.PppoeAccount().Restore(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: result})
	return nil
}
