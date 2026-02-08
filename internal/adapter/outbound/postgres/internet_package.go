package postgres_outbound_adapter

import (
	"mikrops/internal/model"
	outbound_port "mikrops/internal/port/outbound"

	"gorm.io/gorm"
)

type internetPackageAdapter struct {
	db *gorm.DB
}

func NewInternetPackageAdapter(db *gorm.DB) outbound_port.InternetPackageDatabasePort {
	return &internetPackageAdapter{db: db}
}

func (a *internetPackageAdapter) Create(data model.InternetPackageInput) (model.InternetPackage, error) {
	pkg := model.InternetPackage{InternetPackageInput: data}
	result := a.db.Create(&pkg)
	return pkg, result.Error
}

func (a *internetPackageAdapter) FindByFilter(filter model.InternetPackageFilter) ([]model.InternetPackage, error) {
	var packages []model.InternetPackage
	query := a.db.Model(&model.InternetPackage{})
	query = applyInternetPackageFilter(query, filter)

	if filter.WithTenant {
		query = query.Preload("Tenant")
	}

	result := query.Find(&packages)
	return packages, result.Error
}

func (a *internetPackageAdapter) FindByID(id string) (model.InternetPackage, error) {
	var pkg model.InternetPackage
	result := a.db.Preload("Tenant").Where("id = ?", id).First(&pkg)
	return pkg, result.Error
}

func (a *internetPackageAdapter) Update(id string, data model.InternetPackageInput) error {
	return a.db.Model(&model.InternetPackage{}).Where("id = ?", id).Updates(data).Error
}

func (a *internetPackageAdapter) Delete(id string) error {
	return a.db.Where("id = ?", id).Delete(&model.InternetPackage{}).Error
}

func applyInternetPackageFilter(query *gorm.DB, filter model.InternetPackageFilter) *gorm.DB {
	if len(filter.IDs) > 0 {
		query = query.Where("id IN ?", filter.IDs)
	}
	if len(filter.TenantIDs) > 0 {
		query = query.Where("tenant_id IN ?", filter.TenantIDs)
	}
	if len(filter.Types) > 0 {
		query = query.Where("type IN ?", filter.Types)
	}
	if len(filter.BillingCycles) > 0 {
		query = query.Where("billing_cycle IN ?", filter.BillingCycles)
	}
	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}
	return query
}
