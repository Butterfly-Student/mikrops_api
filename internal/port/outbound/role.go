package outbound_port

import "mikrops/internal/model"

//go:generate mockgen -source=role.go -destination=./../../../tests/mocks/port/mock_role.go
type RoleDatabasePort interface {
	FindAll() ([]model.Role, error)
	FindByID(id int) (model.Role, error)
	FindByName(name string) (model.Role, error)
}
