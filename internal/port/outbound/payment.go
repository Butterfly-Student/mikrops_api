package outbound_port

import "mikrops/internal/model"

//go:generate mockgen -source=payment.go -destination=./../../../tests/mocks/port/mock_payment.go
type PaymentDatabasePort interface {
	Create(data model.PaymentInput) (model.Payment, error)
	FindByFilter(filter model.PaymentFilter) ([]model.Payment, error)
	FindByID(id string) (model.Payment, error)
	Update(id string, data model.PaymentInput) error
	Delete(id string) error
}
