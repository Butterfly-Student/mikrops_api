package outbound_port

import "go-template/internal/model"

// MikrotikPort defines methods for interacting with MikroTik RouterOS via API
type MikrotikPort interface {
	// Secret
	CreateSecret(router *model.MikrotikRouter, secret *model.PppoeSecret) error
	UpdateSecret(router *model.MikrotikRouter, secret *model.PppoeSecret) error
	DeleteSecret(router *model.MikrotikRouter, id string) error
	GetSecret(router *model.MikrotikRouter, id string) (*model.PppoeSecret, error)
	ListSecrets(router *model.MikrotikRouter) ([]model.PppoeSecret, error)

	// Profile
	CreateProfile(router *model.MikrotikRouter, profile *model.PppoeProfile) error
	UpdateProfile(router *model.MikrotikRouter, profile *model.PppoeProfile) error
	DeleteProfile(router *model.MikrotikRouter, id string) error
	GetProfile(router *model.MikrotikRouter, id string) (*model.PppoeProfile, error)
	ListProfiles(router *model.MikrotikRouter) ([]model.PppoeProfile, error)

	// Active Sessions
	ListActiveSessions(router *model.MikrotikRouter) ([]model.PppoeActive, error)
	GetActiveSession(router *model.MikrotikRouter, id string) (*model.PppoeActive, error)
	RemoveActiveSession(router *model.MikrotikRouter, id string) error
}

// MikrotikDatabasePort defines methods for managing MikroTik router configurations
type MikrotikDatabasePort interface {
	Create(router *model.MikrotikRouter) error
	FindByID(id uint) (*model.MikrotikRouter, error)
	FindAll() ([]model.MikrotikRouter, error)
	Update(router *model.MikrotikRouter) error
	Delete(id uint) error
}
