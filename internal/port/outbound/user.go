package outbound_port

import "go-template/internal/model"

//go:generate mockgen -source=user.go -destination=./../../../tests/mocks/port/mock_user.go
type UserDatabasePort interface {
	Create(user *model.User) error
	FindByEmail(email string) (*model.User, error)
	FindByID(id uint) (*model.User, error)
	Update(user model.User) error
}
