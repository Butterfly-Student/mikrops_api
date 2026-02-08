package outbound_port

import "mikrops/internal/model"

//go:generate mockgen -source=permission.go -destination=./../../../tests/mocks/port/mock_permission.go
type PermissionDatabasePort interface {
	FindByRoleID(roleID int) ([]model.Permission, error)
	FindAll() ([]model.Permission, error)
}
