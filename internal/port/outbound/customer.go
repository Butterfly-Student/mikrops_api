package outbound_port

import "mikrops/internal/model"

//go:generate mockgen -source=customer.go -destination=./../../../tests/mocks/port/mock_customer.go
type CustomerDatabasePort interface {
	Create(data model.CustomerInput) (model.Customer, error)
	FindByFilter(filter model.CustomerFilter) ([]model.Customer, error)
	FindByID(id string) (model.Customer, error)
	FindByUsername(username string) (model.Customer, error)
	Update(id string, data model.CustomerInput) error
	Delete(id string) error
}

type CustomerCachePort interface {
	Set(data model.Customer) error
	Get(customerID string) (model.Customer, error)
	Delete(customerID string) error
}
