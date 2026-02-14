package system_setting

import (
	"context"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
)

type SystemSettingDomain interface {
	GetSetting(ctx context.Context, key string) (*model.SystemSetting, error)
	ListSettings(ctx context.Context, filter model.SystemSettingFilter) ([]model.SystemSetting, error)
	UpdateSetting(ctx context.Context, key string, value string, userID string) error
	GetPublicSettings(ctx context.Context) ([]model.SystemSetting, error)
}

type domain struct {
	dbPort outbound_port.DatabasePort
}

func NewSystemSettingDomain(
	dbPort outbound_port.DatabasePort,
) SystemSettingDomain {
	return &domain{
		dbPort: dbPort,
	}
}

func (d *domain) GetSetting(ctx context.Context, key string) (*model.SystemSetting, error) {
	return d.dbPort.SystemSetting().FindByKey(key)
}

func (d *domain) ListSettings(ctx context.Context, filter model.SystemSettingFilter) ([]model.SystemSetting, error) {
	if filter.IsEmpty() {
		return d.dbPort.SystemSetting().FindAll()
	}

	if filter.Category != nil {
		return d.dbPort.SystemSetting().FindByCategory(*filter.Category)
	}

	if filter.IsPublic != nil {
		return d.dbPort.SystemSetting().FindPublic()
	}

	return d.dbPort.SystemSetting().FindAll()
}

func (d *domain) UpdateSetting(ctx context.Context, key string, value string, userID string) error {
	return d.dbPort.SystemSetting().UpdateByKey(key, value, userID)
}

func (d *domain) GetPublicSettings(ctx context.Context) ([]model.SystemSetting, error) {
	return d.dbPort.SystemSetting().FindPublic()
}
