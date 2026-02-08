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

type customerAdapter struct {
	domain domain.Domain
}

func NewCustomerAdapter(domain domain.Domain) inbound_port.CustomerHttpPort {
	return &customerAdapter{domain: domain}
}

func (h *customerAdapter) List(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_customer_list")

	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(string)

	filter := model.CustomerFilter{TenantIDs: []string{tid}}

	results, err := h.domain.Customer().FindByFilter(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: results})
	return nil
}

func (h *customerAdapter) Create(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_customer_create")

	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(string)

	var payload model.CustomerInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return nil
	}
	payload.TenantID = tid

	result, err := h.domain.Customer().Create(ctx, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusCreated, model.Response{Success: true, Data: result})
	return nil
}

func (h *customerAdapter) Get(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_customer_get")

	id := c.Param("id")

	result, err := h.domain.Customer().FindByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: result})
	return nil
}

func (h *customerAdapter) Update(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_customer_update")

	id := c.Param("id")

	var payload model.CustomerInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return nil
	}

	err := h.domain.Customer().Update(ctx, id, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true})
	return nil
}

func (h *customerAdapter) Delete(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_customer_delete")

	id := c.Param("id")

	err := h.domain.Customer().Delete(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true})
	return nil
}
