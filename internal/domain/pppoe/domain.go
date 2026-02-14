package pppoe

import (
	"context"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
)

type PppoeDomain interface {
	// Secret Management
	CreateSecret(routerID string, secret model.PppoeSecret) error
	GetSecret(routerID string, id string) (*model.PppoeSecret, error)
	ListSecrets(routerID string) ([]model.PppoeSecret, error)
	UpdateSecret(routerID string, secret model.PppoeSecret) error
	DeleteSecret(routerID string, id string) error

	// Profile Management
	CreateProfile(routerID string, profile model.PppoeProfile) error
	GetProfile(routerID string, id string) (*model.PppoeProfile, error)
	ListProfiles(routerID string) ([]model.PppoeProfile, error)
	UpdateProfile(routerID string, profile model.PppoeProfile) error
	DeleteProfile(routerID string, id string) error

	// Session Management
	ListActiveSessions(routerID string) ([]model.PppoeActive, error)
	ListInactiveSessions(routerID string) ([]model.PppoeSecret, error)
	GetActiveSession(routerID string, id string) (*model.PppoeActive, error)
	RemoveActiveSession(routerID string, id string) error

	// Webhooks
	HandleOnUp(data model.PppoeCallbackData) error
	HandleOnDown(data model.PppoeCallbackData) error

	// Realtime
	SubscribeToEvents(ctx context.Context) (<-chan model.WebSocketMessage, error)
}

type domain struct {
	dbPort       outbound_port.DatabasePort
	cachePort    outbound_port.CachePort
	mikrotikPort outbound_port.MikrotikPort
}

func NewPppoeDomain(
	dbPort outbound_port.DatabasePort,
	cachePort outbound_port.CachePort,
	mikrotikPort outbound_port.MikrotikPort,
) PppoeDomain {
	return &domain{
		dbPort:       dbPort,
		cachePort:    cachePort,
		mikrotikPort: mikrotikPort,
	}
}

func (d *domain) getRouter(routerID string) (*model.MikrotikRouter, error) {
	return d.dbPort.Mikrotik().FindByID(routerID)
}

// Secret Management

func (d *domain) CreateSecret(routerID string, secret model.PppoeSecret) error {
	router, err := d.getRouter(routerID)
	if err != nil {
		return err
	}
	return d.mikrotikPort.CreateSecret(router, &secret)
}

func (d *domain) GetSecret(routerID string, id string) (*model.PppoeSecret, error) {
	router, err := d.getRouter(routerID)
	if err != nil {
		return nil, err
	}
	return d.mikrotikPort.GetSecret(router, id)
}

func (d *domain) ListSecrets(routerID string) ([]model.PppoeSecret, error) {
	router, err := d.getRouter(routerID)
	if err != nil {
		return nil, err
	}
	return d.mikrotikPort.ListSecrets(router)
}

func (d *domain) UpdateSecret(routerID string, secret model.PppoeSecret) error {
	router, err := d.getRouter(routerID)
	if err != nil {
		return err
	}
	return d.mikrotikPort.UpdateSecret(router, &secret)
}

func (d *domain) DeleteSecret(routerID string, id string) error {
	router, err := d.getRouter(routerID)
	if err != nil {
		return err
	}
	return d.mikrotikPort.DeleteSecret(router, id)
}

// Profile Management

func (d *domain) CreateProfile(routerID string, profile model.PppoeProfile) error {
	router, err := d.getRouter(routerID)
	if err != nil {
		return err
	}
	return d.mikrotikPort.CreateProfile(router, &profile)
}

func (d *domain) GetProfile(routerID string, id string) (*model.PppoeProfile, error) {
	router, err := d.getRouter(routerID)
	if err != nil {
		return nil, err
	}
	return d.mikrotikPort.GetProfile(router, id)
}

func (d *domain) ListProfiles(routerID string) ([]model.PppoeProfile, error) {
	router, err := d.getRouter(routerID)
	if err != nil {
		return nil, err
	}
	return d.mikrotikPort.ListProfiles(router)
}

func (d *domain) UpdateProfile(routerID string, profile model.PppoeProfile) error {
	router, err := d.getRouter(routerID)
	if err != nil {
		return err
	}
	return d.mikrotikPort.UpdateProfile(router, &profile)
}

func (d *domain) DeleteProfile(routerID string, id string) error {
	router, err := d.getRouter(routerID)
	if err != nil {
		return err
	}
	return d.mikrotikPort.DeleteProfile(router, id)
}

// Session Management

func (d *domain) ListActiveSessions(routerID string) ([]model.PppoeActive, error) {
	router, err := d.getRouter(routerID)
	if err != nil {
		return nil, err
	}
	return d.mikrotikPort.ListActiveSessions(router)
}

func (d *domain) ListInactiveSessions(routerID string) ([]model.PppoeSecret, error) {
	router, err := d.getRouter(routerID)
	if err != nil {
		return nil, err
	}

	secrets, err := d.mikrotikPort.ListSecrets(router)
	if err != nil {
		return nil, err
	}

	activeSessions, err := d.mikrotikPort.ListActiveSessions(router)
	if err != nil {
		return nil, err
	}

	activeUsers := make(map[string]bool)
	for _, session := range activeSessions {
		activeUsers[session.Name] = true
	}

	var inactiveUsers []model.PppoeSecret
	for _, secret := range secrets {
		if _, isActive := activeUsers[secret.Name]; !isActive && !secret.Disabled {
			inactiveUsers = append(inactiveUsers, secret)
		}
	}

	return inactiveUsers, nil
}

func (d *domain) GetActiveSession(routerID string, id string) (*model.PppoeActive, error) {
	router, err := d.getRouter(routerID)
	if err != nil {
		return nil, err
	}
	return d.mikrotikPort.GetActiveSession(router, id)
}

func (d *domain) RemoveActiveSession(routerID string, id string) error {
	router, err := d.getRouter(routerID)
	if err != nil {
		return err
	}
	return d.mikrotikPort.RemoveActiveSession(router, id)
}

// Webhooks

func (d *domain) HandleOnUp(data model.PppoeCallbackData) error {
	event := model.WebSocketMessage{
		Event: "session_up",
		Data:  data,
	}

	return d.cachePort.PppoePubSub().PublishSessionEvent(event.Event, event.Data)
}

func (d *domain) HandleOnDown(data model.PppoeCallbackData) error {
	event := model.WebSocketMessage{
		Event: "session_down",
		Data:  data,
	}
	return d.cachePort.PppoePubSub().PublishSessionEvent(event.Event, event.Data)
}

// Realtime

func (d *domain) SubscribeToEvents(ctx context.Context) (<-chan model.WebSocketMessage, error) {
	return d.cachePort.PppoePubSub().SubscribeToSessionEvents(ctx)
}
