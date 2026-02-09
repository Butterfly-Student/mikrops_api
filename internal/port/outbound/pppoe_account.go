package outbound_port

import "mikrops/internal/model"

//go:generate mockgen -source=pppoe_account.go -destination=./../../../tests/mocks/port/mock_pppoe_account.go
type PppoeAccountDatabasePort interface {
	Create(data model.PppoeAccountInput) (model.PppoeAccount, error)
	FindByFilter(filter model.PppoeAccountFilter) ([]model.PppoeAccount, error)
	FindByID(id string) (model.PppoeAccount, error)
	Update(id string, data model.PppoeAccountInput) error
	Delete(id string) error
}
