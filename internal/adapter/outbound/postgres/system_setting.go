package postgres_outbound_adapter

import (
	"errors"
	"strconv"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
	"gorm.io/gorm"
)

const tableSystemSetting = "system_settings"

type SystemSettingAdapter struct {
	db *gorm.DB
}

func NewSystemSettingAdapter(db *gorm.DB) outbound_port.SystemSettingDatabasePort {
	return &SystemSettingAdapter{db: db}
}

func (a *SystemSettingAdapter) Create(setting *model.SystemSetting) error {
	model.SystemSettingPrepare(setting)
	return a.db.Create(setting).Error
}

func (a *SystemSettingAdapter) FindByID(id string) (*model.SystemSetting, error) {
	var setting model.SystemSetting
	err := a.db.Where("id = ?", id).First(&setting).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("system setting not found")
		}
		return nil, err
	}
	return &setting, nil
}

func (a *SystemSettingAdapter) FindByKey(key string) (*model.SystemSetting, error) {
	var setting model.SystemSetting
	err := a.db.Where("key = ?", key).First(&setting).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("system setting not found")
		}
		return nil, err
	}
	return &setting, nil
}

func (a *SystemSettingAdapter) FindAll() ([]model.SystemSetting, error) {
	var settings []model.SystemSetting
	err := a.db.Order("category, key").Find(&settings).Error
	return settings, err
}

func (a *SystemSettingAdapter) FindByCategory(category string) ([]model.SystemSetting, error) {
	var settings []model.SystemSetting
	err := a.db.Where("category = ?", category).Order("key").Find(&settings).Error
	return settings, err
}

func (a *SystemSettingAdapter) FindPublic() ([]model.SystemSetting, error) {
	var settings []model.SystemSetting
	err := a.db.Where("is_public = ?", true).Order("category, key").Find(&settings).Error
	return settings, err
}

func (a *SystemSettingAdapter) Update(setting *model.SystemSetting) error {
	return a.db.Save(setting).Error
}

func (a *SystemSettingAdapter) UpdateByKey(key string, value string, userID string) error {
	setting, err := a.FindByKey(key)
	if err != nil {
		return err
	}

	setting.Value = &value
	if userID != "" {
		uid, _ := strconv.ParseUint(userID, 10, 64)
		updatedBy := uint(uid)
		setting.UpdatedBy = &updatedBy
	}

	return a.db.Save(setting).Error
}

func (a *SystemSettingAdapter) Delete(id string) error {
	return a.db.Delete(&model.SystemSetting{}, "id = ?", id).Error
}
