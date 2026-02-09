package postgres_outbound_adapter

import (
	"mikrops/internal/model"
	outbound_port "mikrops/internal/port/outbound"

	"gorm.io/gorm"
)

type customerRegistrationAdapter struct {
	db *gorm.DB
}

func NewCustomerRegistrationAdapter(db *gorm.DB) outbound_port.CustomerRegistrationDatabasePort {
	return &customerRegistrationAdapter{db: db}
}

func (a *customerRegistrationAdapter) Create(data model.CustomerRegistrationInput) (model.CustomerRegistration, error) {
	registration := model.CustomerRegistration{CustomerRegistrationInput: data}
	result := a.db.Create(&registration)
	return registration, result.Error
}

func (a *customerRegistrationAdapter) FindByFilter(filter model.CustomerRegistrationFilter) ([]model.CustomerRegistration, error) {
	var registrations []model.CustomerRegistration
	query := a.db.Model(&model.CustomerRegistration{})
	query = applyCustomerRegistrationFilter(query, filter)

	if filter.WithPackage {
		query = query.Preload("InternetPackage")
	}
	if filter.WithNas {
		query = query.Preload("Nas")
	}
	if filter.WithCustomer {
		query = query.Preload("Customer")
	}

	result := query.Order("created_at DESC").Find(&registrations)
	return registrations, result.Error
}

func (a *customerRegistrationAdapter) FindByID(id string) (model.CustomerRegistration, error) {
	var registration model.CustomerRegistration
	result := a.db.Preload("InternetPackage").Preload("Nas").Preload("Customer").Preload("ApprovedByStaff").Where("id = ?", id).First(&registration)
	return registration, result.Error
}

func (a *customerRegistrationAdapter) Update(id string, data model.CustomerRegistrationInput) error {
	return a.db.Model(&model.CustomerRegistration{}).Where("id = ?", id).Updates(data).Error
}

func (a *customerRegistrationAdapter) Delete(id string) error {
	return a.db.Where("id = ?", id).Delete(&model.CustomerRegistration{}).Error
}

func applyCustomerRegistrationFilter(query *gorm.DB, filter model.CustomerRegistrationFilter) *gorm.DB {
	if len(filter.IDs) > 0 {
		query = query.Where("id IN ?", filter.IDs)
	}
	if len(filter.TenantIDs) > 0 {
		query = query.Where("tenant_id IN ?", filter.TenantIDs)
	}
	if len(filter.Statuses) > 0 {
		query = query.Where("status IN ?", filter.Statuses)
	}
	return query
}
