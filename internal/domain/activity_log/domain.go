package activity_log

import (
	"context"

	"github.com/palantir/stacktrace"

	"mikrops/internal/model"
	outbound_port "mikrops/internal/port/outbound"
)

type ActivityLogDomain interface {
	Create(ctx context.Context, input model.ActivityLogInput) (model.ActivityLog, error)
	FindByFilter(ctx context.Context, filter model.ActivityLogFilter) ([]model.ActivityLog, error)
	FindByID(ctx context.Context, id string) (model.ActivityLog, error)
}

type activityLogDomain struct {
	databasePort outbound_port.DatabasePort
	messagePort  outbound_port.MessagePort
	cachePort    outbound_port.CachePort
	workflowPort outbound_port.WorkflowPort
}

func NewActivityLogDomain(
	databasePort outbound_port.DatabasePort,
	messagePort outbound_port.MessagePort,
	cachePort outbound_port.CachePort,
	workflowPort outbound_port.WorkflowPort,
) ActivityLogDomain {
	return &activityLogDomain{
		databasePort: databasePort,
		messagePort:  messagePort,
		cachePort:    cachePort,
		workflowPort: workflowPort,
	}
}

func (d *activityLogDomain) Create(ctx context.Context, input model.ActivityLogInput) (model.ActivityLog, error) {
	if input.TenantID == "" {
		return model.ActivityLog{}, stacktrace.NewError("tenant_id is required")
	}
	if input.Action == "" {
		return model.ActivityLog{}, stacktrace.NewError("action is required")
	}

	log, err := d.databasePort.ActivityLog().Create(input)
	if err != nil {
		return model.ActivityLog{}, stacktrace.Propagate(err, "failed to create activity log")
	}

	return log, nil
}

func (d *activityLogDomain) FindByFilter(ctx context.Context, filter model.ActivityLogFilter) ([]model.ActivityLog, error) {
	if filter.IsEmpty() {
		return nil, stacktrace.NewError("filter is empty")
	}

	logs, err := d.databasePort.ActivityLog().FindByFilter(filter)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to find activity logs")
	}

	return logs, nil
}

func (d *activityLogDomain) FindByID(ctx context.Context, id string) (model.ActivityLog, error) {
	if id == "" {
		return model.ActivityLog{}, stacktrace.NewError("id is empty")
	}

	log, err := d.databasePort.ActivityLog().FindByID(id)
	if err != nil {
		return model.ActivityLog{}, stacktrace.Propagate(err, "failed to find activity log")
	}

	return log, nil
}
