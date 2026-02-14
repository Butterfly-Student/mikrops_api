package outbound_port

import (
	"go-template/internal/model"
)

type PaymentDatabasePort interface {
	Create(payment *model.Payment) error
	FindByID(id string) (*model.Payment, error)
	Find(filter model.PaymentFilter) ([]model.Payment, error)
	Update(payment *model.Payment) error
	Delete(id string) error
	FindByXenditExternalID(externalID string) (*model.Payment, error)
	FindByCustomerID(customerID string) ([]model.Payment, error)
	FindByInvoiceID(invoiceID string) ([]model.Payment, error)
}

type PaymentAllocationDatabasePort interface {
	Create(allocation *model.PaymentAllocation) error
	FindByID(id string) (*model.PaymentAllocation, error)
	FindByPaymentID(paymentID string) ([]model.PaymentAllocation, error)
	FindByInvoiceID(invoiceID string) ([]model.PaymentAllocation, error)
	Delete(id string) error
}
