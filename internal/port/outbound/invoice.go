package outbound_port

import (
	"time"

	"go-template/internal/model"
)

//go:generate mockgen -source=invoice.go -destination=./../../../tests/mocks/port/mock_invoice.go
type InvoiceDatabasePort interface {
	Create(invoice *model.Invoice) error
	CreateWithItems(invoice *model.Invoice, items []model.InvoiceItem) error
	FindByFilter(filter model.InvoiceFilter) ([]model.Invoice, error)
	Update(invoice *model.Invoice) error
	Delete(id string) error
	GetNextSequenceNumber(year int, month int) (int, error)
	FindOverdueInvoices() ([]model.Invoice, error)
	FindInvoicesDueInDays(days int) ([]model.Invoice, error)
	CountInvoicesByPeriod(year int, month int) (int64, error)
	GetMonthlyRevenue(year int, month int) (map[string]interface{}, error)
	UpdatePaymentStatus(invoiceID string, paidAmount string, paymentDate time.Time, paymentMethod string) error
}
