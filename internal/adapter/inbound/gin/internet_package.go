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

type internetPackageAdapter struct {
	domain domain.Domain
}

func NewInternetPackageAdapter(domain domain.Domain) inbound_port.InternetPackageHttpPort {
	return &internetPackageAdapter{domain: domain}
}

func (h *internetPackageAdapter) List(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_package_list")

	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(string)

	filter := model.InternetPackageFilter{TenantIDs: []string{tid}}

	results, err := h.domain.InternetPackage().FindByFilter(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: results})
	return nil
}

func (h *internetPackageAdapter) Create(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_package_create")

	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(string)

	var payload model.InternetPackageInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return nil
	}
	payload.TenantID = tid

	result, err := h.domain.InternetPackage().Create(ctx, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusCreated, model.Response{Success: true, Data: result})
	return nil
}

func (h *internetPackageAdapter) Get(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_package_get")

	id := c.Param("id")

	result, err := h.domain.InternetPackage().FindByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: result})
	return nil
}

func (h *internetPackageAdapter) Update(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_package_update")

	id := c.Param("id")

	var payload model.InternetPackageInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return nil
	}

	err := h.domain.InternetPackage().Update(ctx, id, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true})
	return nil
}

func (h *internetPackageAdapter) Delete(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_package_delete")

	id := c.Param("id")

	err := h.domain.InternetPackage().Delete(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true})
	return nil
}
