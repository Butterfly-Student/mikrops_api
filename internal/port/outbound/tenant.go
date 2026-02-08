package outbound_port

import "mikrops/internal/model"

//go:generate mockgen -source=tenant.go -destination=./../../../tests/mocks/port/mock_tenant.go
type TenantDatabasePort interface {
	Create(data model.TenantInput) (model.Tenant, error)
	FindByFilter(filter model.TenantFilter) ([]model.Tenant, error)
	FindByID(id string) (model.Tenant, error)
	Update(id string, data model.TenantInput) error
	Delete(id string) error
}
