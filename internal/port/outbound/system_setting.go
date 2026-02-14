package outbound_port

import "go-template/internal/model"

type SystemSettingDatabasePort interface {
	Create(setting *model.SystemSetting) error
	FindByID(id string) (*model.SystemSetting, error)
	FindByKey(key string) (*model.SystemSetting, error)
	FindAll() ([]model.SystemSetting, error)
	FindByCategory(category string) ([]model.SystemSetting, error)
	FindPublic() ([]model.SystemSetting, error)
	Update(setting *model.SystemSetting) error
	UpdateByKey(key string, value string, userID string) error
	Delete(id string) error
}
