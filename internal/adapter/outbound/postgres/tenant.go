package postgres_outbound_adapter

import (
	"mikrops/internal/model"
	outbound_port "mikrops/internal/port/outbound"

	"gorm.io/gorm"
)

type tenantAdapter struct {
	db *gorm.DB
}

func NewTenantAdapter(db *gorm.DB) outbound_port.TenantDatabasePort {
	return &tenantAdapter{db: db}
}

func (a *tenantAdapter) Create(data model.TenantInput) (model.Tenant, error) {
	tenant := model.Tenant{TenantInput: data}
	result := a.db.Create(&tenant)
	return tenant, result.Error
}

func (a *tenantAdapter) FindByFilter(filter model.TenantFilter) ([]model.Tenant, error) {
	var tenants []model.Tenant
	query := a.db.Model(&model.Tenant{})
	query = applyTenantFilter(query, filter)
	result := query.Find(&tenants)
	return tenants, result.Error
}

func (a *tenantAdapter) FindByID(id string) (model.Tenant, error) {
	var tenant model.Tenant
	result := a.db.Where("id = ?", id).First(&tenant)
	return tenant, result.Error
}

func (a *tenantAdapter) Update(id string, data model.TenantInput) error {
	return a.db.Model(&model.Tenant{}).Where("id = ?", id).Updates(data).Error
}

func (a *tenantAdapter) Delete(id string) error {
	return a.db.Where("id = ?", id).Delete(&model.Tenant{}).Error
}

func applyTenantFilter(query *gorm.DB, filter model.TenantFilter) *gorm.DB {
	if len(filter.IDs) > 0 {
		query = query.Where("id IN ?", filter.IDs)
	}
	if len(filter.Slugs) > 0 {
		query = query.Where("slug IN ?", filter.Slugs)
	}
	if len(filter.Emails) > 0 {
		query = query.Where("email IN ?", filter.Emails)
	}
	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}
	if len(filter.SubscriptionPlans) > 0 {
		query = query.Where("subscription_plan IN ?", filter.SubscriptionPlans)
	}
	return query
}
