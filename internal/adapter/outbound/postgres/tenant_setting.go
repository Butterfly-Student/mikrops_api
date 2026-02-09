package postgres_outbound_adapter

import (
	"mikrops/internal/model"
	outbound_port "mikrops/internal/port/outbound"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type tenantSettingAdapter struct {
	db *gorm.DB
}

func NewTenantSettingAdapter(db *gorm.DB) outbound_port.TenantSettingDatabasePort {
	return &tenantSettingAdapter{db: db}
}

func (a *tenantSettingAdapter) Create(data model.TenantSettingInput) (model.TenantSetting, error) {
	setting := model.TenantSetting{TenantSettingInput: data}
	result := a.db.Create(&setting)
	return setting, result.Error
}

func (a *tenantSettingAdapter) FindByFilter(filter model.TenantSettingFilter) ([]model.TenantSetting, error) {
	var settings []model.TenantSetting
	query := a.db.Model(&model.TenantSetting{})
	query = applyTenantSettingFilter(query, filter)
	result := query.Find(&settings)
	return settings, result.Error
}

func (a *tenantSettingAdapter) FindByID(id string) (model.TenantSetting, error) {
	var setting model.TenantSetting
	result := a.db.Where("id = ?", id).First(&setting)
	return setting, result.Error
}

func (a *tenantSettingAdapter) FindByTenantID(tenantID string) (model.TenantSetting, error) {
	var setting model.TenantSetting
	result := a.db.Where("tenant_id = ?", tenantID).First(&setting)
	return setting, result.Error
}

func (a *tenantSettingAdapter) Update(id string, data model.TenantSettingInput) error {
	return a.db.Model(&model.TenantSetting{}).Where("id = ?", id).Updates(data).Error
}

func (a *tenantSettingAdapter) Upsert(data model.TenantSettingInput) (model.TenantSetting, error) {
	setting := model.TenantSetting{TenantSettingInput: data}
	result := a.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "tenant_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"cutoff_day", "grace_period", "isolir_profile_name", "wa_gateway_api", "auto_approve_registration", "default_ppp_password_type", "default_ppp_password_length", "updated_at"}),
	}).Create(&setting)
	return setting, result.Error
}

func (a *tenantSettingAdapter) Delete(id string) error {
	return a.db.Where("id = ?", id).Delete(&model.TenantSetting{}).Error
}

func applyTenantSettingFilter(query *gorm.DB, filter model.TenantSettingFilter) *gorm.DB {
	if len(filter.IDs) > 0 {
		query = query.Where("id IN ?", filter.IDs)
	}
	if len(filter.TenantIDs) > 0 {
		query = query.Where("tenant_id IN ?", filter.TenantIDs)
	}
	return query
}
