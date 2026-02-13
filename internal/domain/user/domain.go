package user

import (
	"errors"
	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
)

type UserDomain interface {
	GetProfile(userID uint) (*model.User, error)
	UpdateProfile(userID uint, req model.UserInput) error
}

type domain struct {
	dbPort outbound_port.DatabasePort
}

func NewUserDomain(dbPort outbound_port.DatabasePort) UserDomain {
	return &domain{
		dbPort: dbPort,
	}
}

func (d *domain) GetProfile(userID uint) (*model.User, error) {
	return d.dbPort.User().FindByID(userID)
}

func (d *domain) UpdateProfile(userID uint, req model.UserInput) error {
	user, err := d.dbPort.User().FindByID(userID)
	if err != nil {
		return err
	}

	if req.Name != "" {
		user.Name = req.Name
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
