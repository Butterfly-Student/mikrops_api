package outbound_port

import "mikrops/internal/model"

//go:generate mockgen -source=tenant_setting.go -destination=./../../../tests/mocks/port/mock_tenant_setting.go
type TenantSettingDatabasePort interface {
	Create(data model.TenantSettingInput) (model.TenantSetting, error)
	FindByFilter(filter model.TenantSettingFilter) ([]model.TenantSetting, error)
	FindByID(id string) (model.TenantSetting, error)
	FindByTenantID(tenantID string) (model.TenantSetting, error)
	Update(id string, data model.TenantSettingInput) error
	Upsert(data model.TenantSettingInput) (model.TenantSetting, error)
	Delete(id string) error
}
