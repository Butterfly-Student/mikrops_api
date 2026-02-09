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

type mikrotikSyncLogAdapter struct {
	domain domain.Domain
}

func NewMikrotikSyncLogAdapter(domain domain.Domain) inbound_port.MikrotikSyncLogHttpPort {
	return &mikrotikSyncLogAdapter{domain: domain}
}

func (h *mikrotikSyncLogAdapter) List(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_mikrotik_sync_log_list")

	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(string)

	filter := model.MikrotikSyncLogFilter{
		TenantIDs: []string{tid},
	}

	nasID := c.Query("nas_id")
	if nasID != "" {
		filter.NasIDs = []string{nasID}
	}

	status := c.Query("status")
	if status != "" {
		filter.Statuses = []string{status}
	}

	results, err := h.domain.MikrotikSyncLog().FindByFilter(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: results})
	return nil
}

func (h *mikrotikSyncLogAdapter) Get(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_mikrotik_sync_log_get")

	id := c.Param("id")

	result, err := h.domain.MikrotikSyncLog().FindByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: result})
	return nil
}
