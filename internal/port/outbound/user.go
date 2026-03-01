package outbound_port

import "go-template/internal/model"

//go:generate mockgen -source=user.go -destination=./../../../tests/mocks/port/mock_user.go
type UserDatabasePort interface {
	Create(user *model.AdminUser) error
	FindByEmail(email string) (*model.AdminUser, error)
	FindByID(id string) (*model.AdminUser, error)
	Update(user model.AdminUser) error
}
