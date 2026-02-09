package tenant_setting

import (
	"context"

	"github.com/palantir/stacktrace"

	"mikrops/internal/model"
	outbound_port "mikrops/internal/port/outbound"
)

type TenantSettingDomain interface {
	Upsert(ctx context.Context, input model.TenantSettingInput) (model.TenantSetting, error)
	FindByTenantID(ctx context.Context, tenantID string) (model.TenantSetting, error)
}

type tenantSettingDomain struct {
	databasePort outbound_port.DatabasePort
	messagePort  outbound_port.MessagePort
	cachePort    outbound_port.CachePort
	workflowPort outbound_port.WorkflowPort
}

func NewTenantSettingDomain(
	databasePort outbound_port.DatabasePort,
	messagePort outbound_port.MessagePort,
	cachePort outbound_port.CachePort,
	workflowPort outbound_port.WorkflowPort,
) TenantSettingDomain {
	return &tenantSettingDomain{
		databasePort: databasePort,
		messagePort:  messagePort,
		cachePort:    cachePort,
		workflowPort: workflowPort,
	}
}

func (d *tenantSettingDomain) Upsert(ctx context.Context, input model.TenantSettingInput) (model.TenantSetting, error) {
	if input.TenantID == "" {
		return model.TenantSetting{}, stacktrace.NewError("tenant_id is required")
	}

	if input.CutoffDay < 1 || input.CutoffDay > 28 {
		input.CutoffDay = 1
	}

	setting, err := d.databasePort.TenantSetting().Upsert(input)
	if err != nil {
		return model.TenantSetting{}, stacktrace.Propagate(err, "failed to upsert tenant setting")
	}

	return setting, nil
}

func (d *tenantSettingDomain) FindByTenantID(ctx context.Context, tenantID string) (model.TenantSetting, error) {
	if tenantID == "" {
		return model.TenantSetting{}, stacktrace.NewError("tenant_id is empty")
	}

	setting, err := d.databasePort.TenantSetting().FindByTenantID(tenantID)
	if err != nil {
		return model.TenantSetting{}, stacktrace.Propagate(err, "failed to find tenant setting")
	}

	return setting, nil
}
