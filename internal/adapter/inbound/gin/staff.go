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

type staffAdapter struct {
	domain domain.Domain
}

func NewStaffAdapter(domain domain.Domain) inbound_port.StaffHttpPort {
	return &staffAdapter{domain: domain}
}

func (h *staffAdapter) List(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_staff_list")

	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(string)

	filter := model.StaffFilter{
		TenantIDs: []string{tid},
		WithRole:  true,
	}

	results, err := h.domain.Staff().FindByFilter(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: results})
	return nil
}

func (h *staffAdapter) Create(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_staff_create")

	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(string)

	var payload model.StaffInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return nil
	}
	payload.TenantID = tid

	result, err := h.domain.Staff().Create(ctx, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusCreated, model.Response{Success: true, Data: result})
	return nil
}

func (h *staffAdapter) Get(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_staff_get")

	id := c.Param("id")

	result, err := h.domain.Staff().FindByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: result})
	return nil
}

func (h *staffAdapter) Update(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_staff_update")

	id := c.Param("id")

	var payload model.StaffInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return nil
	}

	err := h.domain.Staff().Update(ctx, id, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true})
	return nil
}

func (h *staffAdapter) Delete(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_staff_delete")

	id := c.Param("id")

	err := h.domain.Staff().Delete(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true})
	return nil
}
