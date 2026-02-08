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

type tenantAdapter struct {
	domain domain.Domain
}

func NewTenantAdapter(domain domain.Domain) inbound_port.TenantHttpPort {
	return &tenantAdapter{domain: domain}
}

func (h *tenantAdapter) Get(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_tenant_get")

	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(string)

	result, err := h.domain.Tenant().FindByID(ctx, tid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: result})
	return nil
}

func (h *tenantAdapter) Update(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_tenant_update")

	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(string)

	var payload model.TenantInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return nil
	}

	err := h.domain.Tenant().Update(ctx, tid, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true})
	return nil
}
