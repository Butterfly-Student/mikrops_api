package user

import (
	"errors"
	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
)

type UserDomain interface {
	GetProfile(userID string) (*model.AdminUser, error)
	UpdateProfile(userID string, req model.AdminUserInput) error
}

type domain struct {
	dbPort outbound_port.DatabasePort
}

func NewUserDomain(dbPort outbound_port.DatabasePort) UserDomain {
	return &domain{
		dbPort: dbPort,
	}
}

func (d *domain) GetProfile(userID string) (*model.AdminUser, error) {
	return d.dbPort.User().FindByID(userID)
}

func (d *domain) UpdateProfile(userID string, req model.AdminUserInput) error {
	user, err := d.dbPort.User().FindByID(userID)
	if err != nil {
		return err
	}

	if req.FullName != "" {
		user.FullName = req.FullName
	}
	if req.Email != "" {
		// Check if email already exists
		existing, err := d.dbPort.User().FindByEmail(req.Email)
		if err == nil && existing.ID != user.ID {
			return errors.New("email already taken")
		}
		user.Email = req.Email
	}

	return d.dbPort.User().Update(*user)
}
