package postgres_outbound_adapter

import (
	"mikrops/internal/model"
	outbound_port "mikrops/internal/port/outbound"

	"gorm.io/gorm"
)

type paymentAdapter struct {
	db *gorm.DB
}

func NewPaymentAdapter(db *gorm.DB) outbound_port.PaymentDatabasePort {
	return &paymentAdapter{db: db}
}

func (a *paymentAdapter) Create(data model.PaymentInput) (model.Payment, error) {
	payment := model.Payment{PaymentInput: data}
	result := a.db.Create(&payment)
	return payment, result.Error
}

func (a *paymentAdapter) FindByFilter(filter model.PaymentFilter) ([]model.Payment, error) {
	var payments []model.Payment
	query := a.db.Model(&model.Payment{})
	query = applyPaymentFilter(query, filter)

	if filter.WithTenant {
		query = query.Preload("Tenant")
	}
	if filter.WithInvoice {
		query = query.Preload("Invoice")
	}
	if filter.WithPaymentMethod {
		query = query.Preload("PaymentMethod")
	}
	if filter.WithVerifiedBy {
		query = query.Preload("VerifiedBy")
	}

	result := query.Find(&payments)
	return payments, result.Error
}

func (a *paymentAdapter) FindByID(id string) (model.Payment, error) {
	var payment model.Payment
	result := a.db.
		Preload("Tenant").
		Preload("Invoice").
		Preload("PaymentMethod").
		Preload("VerifiedBy").
		Where("id = ?", id).
		First(&payment)
	return payment, result.Error
}

func (a *paymentAdapter) Update(id string, data model.PaymentInput) error {
	return a.db.Model(&model.Payment{}).Where("id = ?", id).Updates(data).Error
}

func (a *paymentAdapter) Delete(id string) error {
	return a.db.Where("id = ?", id).Delete(&model.Payment{}).Error
}

func applyPaymentFilter(query *gorm.DB, filter model.PaymentFilter) *gorm.DB {
	if len(filter.IDs) > 0 {
		query = query.Where("id IN ?", filter.IDs)
	}
	if len(filter.TenantIDs) > 0 {
		query = query.Where("tenant_id IN ?", filter.TenantIDs)
	}
	if len(filter.InvoiceIDs) > 0 {
		query = query.Where("invoice_id IN ?", filter.InvoiceIDs)
	}
	if len(filter.PaymentMethodIDs) > 0 {
		query = query.Where("payment_method_id IN ?", filter.PaymentMethodIDs)
	}
	if len(filter.Statuses) > 0 {
		query = query.Where("status IN ?", filter.Statuses)
	}
	if len(filter.VerifiedByIDs) > 0 {
		query = query.Where("verified_by IN ?", filter.VerifiedByIDs)
	}
	return query
}
