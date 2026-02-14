package outbound_port

import "go-template/internal/model"

//go:generate mockgen -source=customer.go -destination=./../../../tests/mocks/port/mock_customer.go
type CustomerDatabasePort interface {
	Create(customer *model.Customer) error
	FindByFilter(filter model.CustomerFilter) ([]model.Customer, error)
	Update(customer *model.Customer) error
	Delete(id string) error
	FindExpiredCustomers() ([]model.Customer, error)
	FindExpiringCustomers(days int) ([]model.Customer, error)
}
