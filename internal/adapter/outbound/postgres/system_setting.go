package postgres_outbound_adapter

import (
	"gorm.io/gorm"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
)

type systemSettingAdapter struct {
	db *gorm.DB
}

func NewSystemSettingAdapter(
	db *gorm.DB,
) outbound_port.SystemSettingDatabasePort {
	return &systemSettingAdapter{
		db: db,
	}
}

func (adapter *systemSettingAdapter) GetByKey(key string) (*model.SystemSetting, error) {
	var setting model.SystemSetting

	err := adapter.db.Model(&model.SystemSetting{}).
		Preload("UpdatedByUser").
		Where("key = ?", key).
		First(&setting).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &setting, nil
}

func (adapter *systemSettingAdapter) UpdateByKey(key string, value string) error {
	return adapter.db.Model(&model.SystemSetting{}).
		Where("key = ?", key).
		Updates(map[string]interface{}{
			"value":      value,
			"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
		}).Error
}

func (adapter *systemSettingAdapter) GetAll(category string) ([]model.SystemSetting, error) {
	var settings []model.SystemSetting

	query := adapter.db.Model(&model.SystemSetting{}).Preload("UpdatedByUser")

	if category != "" {
		query = query.Where("category = ?", category)
	}

	query = query.Order("category ASC, key ASC")

	err := query.Find(&settings).Error
	if err != nil {
		return nil, err
	}

	return settings, nil
}

func (adapter *systemSettingAdapter) Create(setting *model.SystemSetting) error {
	return adapter.db.Create(setting).Error
}

func (adapter *systemSettingAdapter) Update(setting *model.SystemSetting) error {
	return adapter.db.Save(setting).Error
}
