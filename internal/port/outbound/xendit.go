package outbound_port

import (
	"context"

	"mikrops/internal/model"
)

//go:generate mockgen -source=xendit.go -destination=./../../../tests/mocks/port/mock_xendit.go
type XenditPort interface {
	CreateInvoice(ctx context.Context, invoice model.Invoice, customer model.Customer) (externalID string, invoiceURL string, err error)
	CreateVirtualAccount(ctx context.Context, invoice model.Invoice, customer model.Customer, bankCode string) (accountNumber string, err error)
	GetInvoiceStatus(ctx context.Context, externalID string) (status string, err error)
}
