package postgres_outbound_adapter

import (
	"mikrops/internal/model"
	outbound_port "mikrops/internal/port/outbound"

	"gorm.io/gorm"
)

type invoiceAdapter struct {
	db *gorm.DB
}

func NewInvoiceAdapter(db *gorm.DB) outbound_port.InvoiceDatabasePort {
	return &invoiceAdapter{db: db}
}

func (a *invoiceAdapter) Create(data model.InvoiceInput) (model.Invoice, error) {
	invoice := model.Invoice{InvoiceInput: data}
	result := a.db.Create(&invoice)
	return invoice, result.Error
}

func (a *invoiceAdapter) FindByFilter(filter model.InvoiceFilter) ([]model.Invoice, error) {
	var invoices []model.Invoice
	query := a.db.Model(&model.Invoice{})
	query = applyInvoiceFilter(query, filter)

	if filter.WithTenant {
		query = query.Preload("Tenant")
	}
	if filter.WithCustomer {
		query = query.Preload("Customer")
	}
	if filter.WithSubscription {
		query = query.Preload("Subscription")
	}

	result := query.Find(&invoices)
	return invoices, result.Error
}

func (a *invoiceAdapter) FindByID(id string) (model.Invoice, error) {
	var invoice model.Invoice
	result := a.db.
		Preload("Tenant").
		Preload("Customer").
		Preload("Subscription").
		Where("id = ?", id).
		First(&invoice)
	return invoice, result.Error
}

func (a *invoiceAdapter) Update(id string, data model.InvoiceInput) error {
	return a.db.Model(&model.Invoice{}).Where("id = ?", id).Updates(data).Error
}

func (a *invoiceAdapter) BulkCreate(datas []model.InvoiceInput) error {
	invoices := make([]model.Invoice, len(datas))
	for i, d := range datas {
		invoices[i] = model.Invoice{InvoiceInput: d}
	}
	result := a.db.Create(&invoices)
	return result.Error
}

func (a *invoiceAdapter) Delete(id string) error {
	return a.db.Where("id = ?", id).Delete(&model.Invoice{}).Error
}

func applyInvoiceFilter(query *gorm.DB, filter model.InvoiceFilter) *gorm.DB {
	if len(filter.IDs) > 0 {
		query = query.Where("id IN ?", filter.IDs)
	}
	if len(filter.TenantIDs) > 0 {
		query = query.Where("tenant_id IN ?", filter.TenantIDs)
	}
	if len(filter.CustomerIDs) > 0 {
		query = query.Where("customer_id IN ?", filter.CustomerIDs)
	}
	if len(filter.SubscriptionIDs) > 0 {
		query = query.Where("subscription_id IN ?", filter.SubscriptionIDs)
	}
	if len(filter.Statuses) > 0 {
		query = query.Where("status IN ?", filter.Statuses)
	}
	if len(filter.InvoiceNumbers) > 0 {
		query = query.Where("invoice_number IN ?", filter.InvoiceNumbers)
	}
	return query
}
