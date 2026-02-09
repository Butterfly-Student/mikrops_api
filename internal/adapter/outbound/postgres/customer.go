package postgres_outbound_adapter

import (
	"mikrops/internal/model"
	outbound_port "mikrops/internal/port/outbound"

	"gorm.io/gorm"
)

type customerAdapter struct {
	db *gorm.DB
}

func NewCustomerAdapter(db *gorm.DB) outbound_port.CustomerDatabasePort {
	return &customerAdapter{db: db}
}

func (a *customerAdapter) Create(data model.CustomerInput) (model.Customer, error) {
	customer := model.Customer{CustomerInput: data}
	result := a.db.Create(&customer)
	return customer, result.Error
}

func (a *customerAdapter) FindByFilter(filter model.CustomerFilter) ([]model.Customer, error) {
	var customers []model.Customer
	query := a.db.Model(&model.Customer{})
	query = applyCustomerFilter(query, filter)

	if filter.WithTenant {
		query = query.Preload("Tenant")
	}
	if filter.WithNas {
		query = query.Preload("Nas")
	}

	result := query.Find(&customers)
	return customers, result.Error
}

func (a *customerAdapter) FindByID(id string) (model.Customer, error) {
	var customer model.Customer
	result := a.db.Preload("Tenant").Preload("Nas").Where("id = ?", id).First(&customer)
	return customer, result.Error
}

func (a *customerAdapter) FindByUsername(username string) (model.Customer, error) {
	var customer model.Customer
	result := a.db.Preload("Tenant").Preload("Nas").Where("username = ?", username).First(&customer)
	return customer, result.Error
}

func (a *customerAdapter) Update(id string, data model.CustomerInput) error {
	return a.db.Model(&model.Customer{}).Where("id = ?", id).Updates(data).Error
}

func (a *customerAdapter) Delete(id string) error {
	return a.db.Where("id = ?", id).Delete(&model.Customer{}).Error
}

func applyCustomerFilter(query *gorm.DB, filter model.CustomerFilter) *gorm.DB {
	if len(filter.IDs) > 0 {
		query = query.Where("id IN ?", filter.IDs)
	}
	if len(filter.TenantIDs) > 0 {
		query = query.Where("tenant_id IN ?", filter.TenantIDs)
	}
	if len(filter.NasIDs) > 0 {
		query = query.Where("nas_id IN ?", filter.NasIDs)
	}
	if len(filter.Emails) > 0 {
		query = query.Where("email IN ?", filter.Emails)
	}
	if len(filter.Usernames) > 0 {
		query = query.Where("username IN ?", filter.Usernames)
	}
	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}
	if filter.AutoCutoff != nil {
		query = query.Where("auto_cutoff = ?", *filter.AutoCutoff)
	}
	return query
}
