package outbound_port

import "go-template/internal/model"

type CustomerDatabasePort interface {
	Create(customer *model.Customer) error
	FindByID(id string) (*model.Customer, error)
	FindByCustomerCode(code string) (*model.Customer, error)
	FindByRouterID(routerID string) ([]model.Customer, error)
	FindByProfileID(profileID string) ([]model.Customer, error)
	FindAll() ([]model.Customer, error)
	Find(filter model.CustomerFilter) ([]model.Customer, error)
	FindExpired() ([]model.Customer, error)
	FindExpiringSoon(days int) ([]model.Customer, error)
	Update(customer *model.Customer) error
	Delete(id string) error
	FindByPppSecretName(name string) (*model.Customer, error)
}
