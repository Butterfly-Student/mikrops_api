package postgres_outbound_adapter

import (
	"mikrops/internal/model"
	outbound_port "mikrops/internal/port/outbound"

	"gorm.io/gorm"
)

type nasAdapter struct {
	db *gorm.DB
}

func NewNasAdapter(db *gorm.DB) outbound_port.NasDatabasePort {
	return &nasAdapter{db: db}
}

func (a *nasAdapter) Create(data model.NasInput) (model.Nas, error) {
	nas := model.Nas{NasInput: data}
	result := a.db.Create(&nas)
	return nas, result.Error
}

func (a *nasAdapter) FindByFilter(filter model.NasFilter) ([]model.Nas, error) {
	var nasList []model.Nas
	query := a.db.Model(&model.Nas{})
	query = applyNasFilter(query, filter)

	if filter.WithTenant {
		query = query.Preload("Tenant")
	}

	result := query.Find(&nasList)
	return nasList, result.Error
}

func (a *nasAdapter) FindByID(id string) (model.Nas, error) {
	var nas model.Nas
	result := a.db.Preload("Tenant").Where("id = ?", id).First(&nas)
	return nas, result.Error
}

func (a *nasAdapter) Update(id string, data model.NasInput) error {
	return a.db.Model(&model.Nas{}).Where("id = ?", id).Updates(data).Error
}

func (a *nasAdapter) Delete(id string) error {
	return a.db.Where("id = ?", id).Delete(&model.Nas{}).Error
}

func (a *nasAdapter) CountByTenantID(tenantID string) (int, error) {
	var count int64
	result := a.db.Model(&model.Nas{}).Where("tenant_id = ?", tenantID).Count(&count)
	return int(count), result.Error
}

func applyNasFilter(query *gorm.DB, filter model.NasFilter) *gorm.DB {
	if len(filter.IDs) > 0 {
		query = query.Where("id IN ?", filter.IDs)
	}
	if len(filter.TenantIDs) > 0 {
		query = query.Where("tenant_id IN ?", filter.TenantIDs)
	}
	if len(filter.Hosts) > 0 {
		query = query.Where("host IN ?", filter.Hosts)
	}
	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}
	return query
}
