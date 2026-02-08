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

type nasAdapter struct {
	domain domain.Domain
}

func NewNasAdapter(domain domain.Domain) inbound_port.NasHttpPort {
	return &nasAdapter{domain: domain}
}

func (h *nasAdapter) List(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_nas_list")

	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(string)

	filter := model.NasFilter{TenantIDs: []string{tid}}

	results, err := h.domain.Nas().FindByFilter(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: results})
	return nil
}

func (h *nasAdapter) Create(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_nas_create")

	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(string)

	var payload model.NasInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return nil
	}
	payload.TenantID = tid

	result, err := h.domain.Nas().Create(ctx, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusCreated, model.Response{Success: true, Data: result})
	return nil
}

func (h *nasAdapter) Get(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_nas_get")

	id := c.Param("id")

	result, err := h.domain.Nas().FindByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: result})
	return nil
}

func (h *nasAdapter) Update(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_nas_update")

	id := c.Param("id")

	var payload model.NasInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return nil
	}

	err := h.domain.Nas().Update(ctx, id, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true})
	return nil
}

func (h *nasAdapter) Delete(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_nas_delete")

	id := c.Param("id")

	err := h.domain.Nas().Delete(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true})
	return nil
}

func (h *nasAdapter) TestConnection(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_nas_test_connection")

	id := c.Param("id")

	err := h.domain.Nas().TestConnection(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: gin.H{"message": "Connection successful"}})
	return nil
}
