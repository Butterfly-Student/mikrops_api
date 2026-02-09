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

type tenantSettingAdapter struct {
	domain domain.Domain
}

func NewTenantSettingAdapter(domain domain.Domain) inbound_port.TenantSettingHttpPort {
	return &tenantSettingAdapter{domain: domain}
}

func (h *tenantSettingAdapter) Get(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_tenant_setting_get")

	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(string)

	result, err := h.domain.TenantSetting().FindByTenantID(ctx, tid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: result})
	return nil
}

func (h *tenantSettingAdapter) Upsert(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_tenant_setting_upsert")

	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(string)

	var payload model.TenantSettingInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Success: false, Error: err.Error()})
		return nil
	}
	payload.TenantID = tid

	result, err := h.domain.TenantSetting().Upsert(ctx, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: result})
	return nil
}
