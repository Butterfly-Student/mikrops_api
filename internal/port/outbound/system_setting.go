package outbound_port

import "go-template/internal/model"

//go:generate mockgen -source=system_setting.go -destination=./../../../tests/mocks/port/mock_system_setting.go
type SystemSettingDatabasePort interface {
	GetByKey(key string) (*model.SystemSetting, error)
	UpdateByKey(key string, value string) error
	GetAll(category string) ([]model.SystemSetting, error)
	Create(setting *model.SystemSetting) error
	Update(setting *model.SystemSetting) error
}
