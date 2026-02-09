package mikrotik_sync_log

import (
	"context"

	"github.com/palantir/stacktrace"

	"mikrops/internal/model"
	outbound_port "mikrops/internal/port/outbound"
)

type MikrotikSyncLogDomain interface {
	Create(ctx context.Context, input model.MikrotikSyncLogInput) (model.MikrotikSyncLog, error)
	FindByFilter(ctx context.Context, filter model.MikrotikSyncLogFilter) ([]model.MikrotikSyncLog, error)
	FindByID(ctx context.Context, id string) (model.MikrotikSyncLog, error)
}

type mikrotikSyncLogDomain struct {
	databasePort outbound_port.DatabasePort
	messagePort  outbound_port.MessagePort
	cachePort    outbound_port.CachePort
	workflowPort outbound_port.WorkflowPort
}

func NewMikrotikSyncLogDomain(
	databasePort outbound_port.DatabasePort,
	messagePort outbound_port.MessagePort,
	cachePort outbound_port.CachePort,
	workflowPort outbound_port.WorkflowPort,
) MikrotikSyncLogDomain {
	return &mikrotikSyncLogDomain{
		databasePort: databasePort,
		messagePort:  messagePort,
		cachePort:    cachePort,
		workflowPort: workflowPort,
	}
}

func (d *mikrotikSyncLogDomain) Create(ctx context.Context, input model.MikrotikSyncLogInput) (model.MikrotikSyncLog, error) {
	if input.TenantID == "" {
		return model.MikrotikSyncLog{}, stacktrace.NewError("tenant_id is required")
	}
	if input.NasID == "" {
		return model.MikrotikSyncLog{}, stacktrace.NewError("nas_id is required")
	}

	log, err := d.databasePort.MikrotikSyncLog().Create(input)
	if err != nil {
		return model.MikrotikSyncLog{}, stacktrace.Propagate(err, "failed to create mikrotik sync log")
	}

	return log, nil
}

func (d *mikrotikSyncLogDomain) FindByFilter(ctx context.Context, filter model.MikrotikSyncLogFilter) ([]model.MikrotikSyncLog, error) {
	if filter.IsEmpty() {
		return nil, stacktrace.NewError("filter is empty")
	}

	logs, err := d.databasePort.MikrotikSyncLog().FindByFilter(filter)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to find mikrotik sync logs")
	}

	return logs, nil
}

func (d *mikrotikSyncLogDomain) FindByID(ctx context.Context, id string) (model.MikrotikSyncLog, error) {
	if id == "" {
		return model.MikrotikSyncLog{}, stacktrace.NewError("id is empty")
	}

	log, err := d.databasePort.MikrotikSyncLog().FindByID(id)
	if err != nil {
		return model.MikrotikSyncLog{}, stacktrace.Propagate(err, "failed to find mikrotik sync log")
	}

	return log, nil
}
