package mikrotik

import (
	"errors"

	"go-template/internal/model"
)

var (
	ErrRouterNotFound      = errors.New("router not found")
	ErrRouterInvalid       = errors.New("invalid router data")
	ErrNoActiveRouter      = errors.New("no active router found")
	ErrConnectionFailed    = errors.New("failed to connect to router")
	ErrRouterAlreadyActive = errors.New("router is already active")
)

type MikrotikDomain interface {
	Create(name, address, username, password string) (*model.MikrotikRouter, error)
	FindByID(id string) (*model.MikrotikRouter, error)
	FindAll() ([]model.MikrotikRouter, error)
	Update(id string, updates map[string]interface{}) (*model.MikrotikRouter, error)
	Delete(id string) error

	SetActive(id string) (*model.MikrotikRouter, error)
	GetActive() (*model.MikrotikRouter, error)
	DeactivateAll() error

	UpdateLastSeen(id string) error
	TestConnection(id string) error
}
