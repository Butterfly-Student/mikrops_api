package outbound_port

import "mikrops/internal/model"

//go:generate mockgen -source=invoice.go -destination=./../../../tests/mocks/port/mock_invoice.go
type InvoiceDatabasePort interface {
	Create(data model.InvoiceInput) (model.Invoice, error)
	FindByFilter(filter model.InvoiceFilter) ([]model.Invoice, error)
	FindByID(id string) (model.Invoice, error)
	Update(id string, data model.InvoiceInput) error
	Delete(id string) error
	BulkCreate(datas []model.InvoiceInput) error
}
