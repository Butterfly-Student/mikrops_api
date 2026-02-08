package tenant

import (
	"context"

	"github.com/palantir/stacktrace"

	"mikrops/internal/model"
	outbound_port "mikrops/internal/port/outbound"
)

type TenantDomain interface {
	Create(ctx context.Context, input model.TenantInput) (model.Tenant, error)
	FindByFilter(ctx context.Context, filter model.TenantFilter) ([]model.Tenant, error)
	FindByID(ctx context.Context, id string) (model.Tenant, error)
	Update(ctx context.Context, id string, input model.TenantInput) error
	Delete(ctx context.Context, id string) error
}

type tenantDomain struct {
	databasePort outbound_port.DatabasePort
	messagePort  outbound_port.MessagePort
	cachePort    outbound_port.CachePort
	workflowPort outbound_port.WorkflowPort
}

func NewTenantDomain(
	databasePort outbound_port.DatabasePort,
	messagePort outbound_port.MessagePort,
	cachePort outbound_port.CachePort,
	workflowPort outbound_port.WorkflowPort,
) TenantDomain {
	return &tenantDomain{
		databasePort: databasePort,
		messagePort:  messagePort,
		cachePort:    cachePort,
		workflowPort: workflowPort,
	}
}

func (d *tenantDomain) Create(ctx context.Context, input model.TenantInput) (model.Tenant, error) {
	// Validate slug uniqueness
	if input.Slug != "" {
		existing, err := d.databasePort.Tenant().FindByFilter(model.TenantFilter{Slugs: []string{input.Slug}})
		if err != nil {
			return model.Tenant{}, stacktrace.Propagate(err, "failed to check slug uniqueness")
		}
		if len(existing) > 0 {
			return model.Tenant{}, stacktrace.NewError("slug already exists")
		}
	}

	model.TenantPrepare(&input)

	tenant, err := d.databasePort.Tenant().Create(input)
	if err != nil {
		return model.Tenant{}, stacktrace.Propagate(err, "failed to create tenant")
	}

	return tenant, nil
}

func (d *tenantDomain) FindByFilter(ctx context.Context, filter model.TenantFilter) ([]model.Tenant, error) {
	if filter.IsEmpty() {
		return nil, stacktrace.NewError("filter is empty")
	}

	tenants, err := d.databasePort.Tenant().FindByFilter(filter)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to find tenants by filter")
	}

	return tenants, nil
}

func (d *tenantDomain) FindByID(ctx context.Context, id string) (model.Tenant, error) {
	if id == "" {
		return model.Tenant{}, stacktrace.NewError("id is empty")
	}

	tenant, err := d.databasePort.Tenant().FindByID(id)
	if err != nil {
		return model.Tenant{}, stacktrace.Propagate(err, "failed to find tenant by id")
	}

	return tenant, nil
}

func (d *tenantDomain) Update(ctx context.Context, id string, input model.TenantInput) error {
	if id == "" {
		return stacktrace.NewError("id is empty")
	}

	// Check if slug is being changed and validate uniqueness
	if input.Slug != "" {
		existing, err := d.databasePort.Tenant().FindByFilter(model.TenantFilter{Slugs: []string{input.Slug}})
		if err != nil {
			return stacktrace.Propagate(err, "failed to check slug uniqueness")
		}
		// Allow update if no existing tenant found, or if existing tenant is the same one being updated
		if len(existing) > 0 && existing[0].ID != id {
			return stacktrace.NewError("slug already exists")
		}
	}

	err := d.databasePort.Tenant().Update(id, input)
	if err != nil {
		return stacktrace.Propagate(err, "failed to update tenant")
	}

	return nil
}

func (d *tenantDomain) Delete(ctx context.Context, id string) error {
	if id == "" {
		return stacktrace.NewError("id is empty")
	}

	err := d.databasePort.Tenant().Delete(id)
	if err != nil {
		return stacktrace.Propagate(err, "failed to delete tenant")
	}

	return nil
}
