package outbound_port

import (
	"context"

	"go-template/internal/model"
)

//go:generate mockgen -source=payment.go -destination=./../../../tests/mocks/port/mock_payment.go
type PaymentDatabasePort interface {
	Create(ctx context.Context, payment *model.Payment) error
	FindByID(ctx context.Context, id string) (*model.Payment, error)
	FindByNumber(ctx context.Context, number string) (*model.Payment, error)
	FindAll(ctx context.Context, filter *model.PaymentFilter) ([]model.Payment, error)
	Update(ctx context.Context, payment *model.Payment) error
	Delete(ctx context.Context, id string) error
	CreateAllocation(ctx context.Context, allocation *model.PaymentAllocation) error
	GetLastPaymentNumber(ctx context.Context, year, month int) (string, error)
}
