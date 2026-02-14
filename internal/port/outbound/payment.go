package outbound_port

import "go-template/internal/model"

//go:generate mockgen -source=payment.go -destination=./../../../tests/mocks/port/mock_payment.go
type PaymentDatabasePort interface {
	Create(payment *model.Payment) error
	CreateWithAllocations(payment *model.Payment, allocations []model.PaymentAllocation) error
	FindByFilter(filter model.PaymentFilter) ([]model.Payment, error)
	Update(payment *model.Payment) error
	Delete(id string) error
	GetNextSequenceNumber(year int, month int) (int, error)
	FindByTransactionReference(ref string) (*model.Payment, error)
	FindPendingPayments() ([]model.Payment, error)
	AllocateToInvoice(paymentID string, invoiceID string, amount string) error
	GetPaymentAllocations(paymentID string) ([]model.PaymentAllocation, error)
}
