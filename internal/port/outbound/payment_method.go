package outbound_port

import "mikrops/internal/model"

//go:generate mockgen -source=payment_method.go -destination=./../../../tests/mocks/port/mock_payment_method.go
type PaymentMethodDatabasePort interface {
	Create(data model.PaymentMethodInput) (model.PaymentMethod, error)
	FindByFilter(filter model.PaymentMethodFilter) ([]model.PaymentMethod, error)
	FindByID(id string) (model.PaymentMethod, error)
	Update(id string, data model.PaymentMethodInput) error
	Delete(id string) error
}
