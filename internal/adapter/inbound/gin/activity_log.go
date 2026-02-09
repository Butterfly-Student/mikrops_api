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

type activityLogAdapter struct {
	domain domain.Domain
}

func NewActivityLogAdapter(domain domain.Domain) inbound_port.ActivityLogHttpPort {
	return &activityLogAdapter{domain: domain}
}

func (h *activityLogAdapter) List(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_activity_log_list")

	tenantID, _ := c.Get("tenant_id")
	tid, _ := tenantID.(string)

	filter := model.ActivityLogFilter{
		TenantIDs: []string{tid},
	}

	action := c.Query("action")
	if action != "" {
		filter.Actions = []string{action}
	}

	resourceType := c.Query("resource_type")
	if resourceType != "" {
		filter.ResourceTypes = []string{resourceType}
	}

	results, err := h.domain.ActivityLog().FindByFilter(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: results})
	return nil
}

func (h *activityLogAdapter) Get(a any) error {
	c := a.(*gin.Context)
	ctx := activity.NewContext("http_activity_log_get")

	id := c.Param("id")

	result, err := h.domain.ActivityLog().FindByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Success: false, Error: stacktrace.RootCause(err).Error()})
		return nil
	}

	c.JSON(http.StatusOK, model.Response{Success: true, Data: result})
	return nil
}
