package outbound_port

import "mikrops/internal/model"

//go:generate mockgen -source=customer_registration.go -destination=./../../../tests/mocks/port/mock_customer_registration.go
type CustomerRegistrationDatabasePort interface {
	Create(data model.CustomerRegistrationInput) (model.CustomerRegistration, error)
	FindByFilter(filter model.CustomerRegistrationFilter) ([]model.CustomerRegistration, error)
	FindByID(id string) (model.CustomerRegistration, error)
	Update(id string, data model.CustomerRegistrationInput) error
	Delete(id string) error
}
