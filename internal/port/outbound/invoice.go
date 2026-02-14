package outbound_port

import (
	"go-template/internal/model"
)

type InvoiceDatabasePort interface {
	Create(invoice *model.Invoice) error
	FindByID(id string) (*model.Invoice, error)
	FindAll() ([]model.Invoice, error)
	Find(filter model.InvoiceFilter) ([]model.Invoice, error)
	Update(invoice *model.Invoice) error
	Delete(id string) error
	FindByCustomerID(customerID string) ([]model.Invoice, error)
	FindByInvoiceNumber(number string) (*model.Invoice, error)
	FindOverdue() ([]model.Invoice, error)
	FindByBillingPeriod(year int, month int) ([]model.Invoice, error)
}

type InvoiceItemDatabasePort interface {
	Create(item *model.InvoiceItem) error
	FindByID(id string) (*model.InvoiceItem, error)
	FindByInvoiceID(invoiceID string) ([]model.InvoiceItem, error)
	Update(item *model.InvoiceItem) error
	Delete(id string) error
}
