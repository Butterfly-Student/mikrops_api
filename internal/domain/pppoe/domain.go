package pppoe

import (
	"context"

	"go-template/internal/model"
)

type PppoeDomain interface {
	// Secret Management
	CreateSecret(routerID uint, secret model.PppoeSecret) error
	GetSecret(routerID uint, id string) (*model.PppoeSecret, error)
	ListSecrets(routerID uint) ([]model.PppoeSecret, error)
	UpdateSecret(routerID uint, secret model.PppoeSecret) error
	DeleteSecret(routerID uint, id string) error

	// Profile Management
	CreateProfile(routerID uint, profile model.PppoeProfile) error
	GetProfile(routerID uint, id string) (*model.PppoeProfile, error)
	ListProfiles(routerID uint) ([]model.PppoeProfile, error)
	UpdateProfile(routerID uint, profile model.PppoeProfile) error
	DeleteProfile(routerID uint, id string) error

	// Session Management
	ListActiveSessions(routerID uint) ([]model.PppoeActive, error)
	GetActiveSession(routerID uint, id string) (*model.PppoeActive, error)
	RemoveActiveSession(routerID uint, id string) error

	// Webhooks
	HandleOnUp(data model.PppoeCallbackData) error
	HandleOnDown(data model.PppoeCallbackData) error

	// Realtime
	SubscribeToEvents(ctx context.Context) (<-chan model.WebSocketMessage, error)
}
