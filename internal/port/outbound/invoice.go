package outbound_port

import (
	"context"

	"go-template/internal/model"
)

type InvoiceDatabasePort interface {
	Create(ctx context.Context, invoice *model.Invoice) error
	FindByID(ctx context.Context, id string) (*model.Invoice, error)
	FindByNumber(ctx context.Context, number string) (*model.Invoice, error)
	FindAll(ctx context.Context, filter *model.InvoiceFilter) ([]model.Invoice, error)
	Update(ctx context.Context, invoice *model.Invoice) error
	Delete(ctx context.Context, id string) error
	GetLastInvoiceNumber(ctx context.Context, year, month int) (string, error)
	FindOverdueInvoices(ctx context.Context) ([]model.Invoice, error)
}
