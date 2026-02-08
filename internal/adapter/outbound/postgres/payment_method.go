package postgres_outbound_adapter

import (
	"mikrops/internal/model"
	outbound_port "mikrops/internal/port/outbound"

	"gorm.io/gorm"
)

type paymentMethodAdapter struct {
	db *gorm.DB
}

func NewPaymentMethodAdapter(db *gorm.DB) outbound_port.PaymentMethodDatabasePort {
	return &paymentMethodAdapter{db: db}
}

func (a *paymentMethodAdapter) Create(data model.PaymentMethodInput) (model.PaymentMethod, error) {
	paymentMethod := model.PaymentMethod{PaymentMethodInput: data}
	result := a.db.Create(&paymentMethod)
	return paymentMethod, result.Error
}

func (a *paymentMethodAdapter) FindByFilter(filter model.PaymentMethodFilter) ([]model.PaymentMethod, error) {
	var paymentMethods []model.PaymentMethod
	query := a.db.Model(&model.PaymentMethod{})
	query = applyPaymentMethodFilter(query, filter)

	if filter.WithTenant {
		query = query.Preload("Tenant")
	}

	result := query.Find(&paymentMethods)
	return paymentMethods, result.Error
}

func (a *paymentMethodAdapter) FindByID(id string) (model.PaymentMethod, error) {
	var paymentMethod model.PaymentMethod
	result := a.db.Preload("Tenant").Where("id = ?", id).First(&paymentMethod)
	return paymentMethod, result.Error
}

func (a *paymentMethodAdapter) Update(id string, data model.PaymentMethodInput) error {
	return a.db.Model(&model.PaymentMethod{}).Where("id = ?", id).Updates(data).Error
}

func (a *paymentMethodAdapter) Delete(id string) error {
	return a.db.Where("id = ?", id).Delete(&model.PaymentMethod{}).Error
}

func applyPaymentMethodFilter(query *gorm.DB, filter model.PaymentMethodFilter) *gorm.DB {
	if len(filter.IDs) > 0 {
		query = query.Where("id IN ?", filter.IDs)
	}
	if len(filter.TenantIDs) > 0 {
		query = query.Where("tenant_id IN ?", filter.TenantIDs)
	}
	if len(filter.Types) > 0 {
		query = query.Where("type IN ?", filter.Types)
	}
	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}
	return query
}
