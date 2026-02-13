package pppoe

import (
	"context"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
)

type service struct {
	dbPort       outbound_port.DatabasePort
	cachePort    outbound_port.CachePort
	mikrotikPort outbound_port.MikrotikPort
}

func NewPppoeDomain(
	dbPort outbound_port.DatabasePort,
	cachePort outbound_port.CachePort,
	mikrotikPort outbound_port.MikrotikPort,
) PppoeDomain {
	return &service{
		dbPort:       dbPort,
		cachePort:    cachePort,
		mikrotikPort: mikrotikPort,
	}
}

func (s *service) getRouter(routerID uint) (*model.MikrotikRouter, error) {
	return s.dbPort.Mikrotik().FindByID(routerID)
}

// Secret Management

func (s *service) CreateSecret(routerID uint, secret model.PppoeSecret) error {
	router, err := s.getRouter(routerID)
	if err != nil {
		return err
	}
	return s.mikrotikPort.CreateSecret(router, &secret)
}

func (s *service) GetSecret(routerID uint, id string) (*model.PppoeSecret, error) {
	router, err := s.getRouter(routerID)
	if err != nil {
		return nil, err
	}
	return s.mikrotikPort.GetSecret(router, id)
}

func (s *service) ListSecrets(routerID uint) ([]model.PppoeSecret, error) {
	router, err := s.getRouter(routerID)
	if err != nil {
		return nil, err
	}
	return s.mikrotikPort.ListSecrets(router)
}

func (s *service) UpdateSecret(routerID uint, secret model.PppoeSecret) error {
	router, err := s.getRouter(routerID)
	if err != nil {
		return err
	}
	return s.mikrotikPort.UpdateSecret(router, &secret)
}

func (s *service) DeleteSecret(routerID uint, id string) error {
	router, err := s.getRouter(routerID)
	if err != nil {
		return err
	}
	return s.mikrotikPort.DeleteSecret(router, id)
}

// Profile Management

func (s *service) CreateProfile(routerID uint, profile model.PppoeProfile) error {
	router, err := s.getRouter(routerID)
	if err != nil {
		return err
	}
	return s.mikrotikPort.CreateProfile(router, &profile)
}

func (s *service) GetProfile(routerID uint, id string) (*model.PppoeProfile, error) {
	router, err := s.getRouter(routerID)
	if err != nil {
		return nil, err
	}
	return s.mikrotikPort.GetProfile(router, id)
}

func (s *service) ListProfiles(routerID uint) ([]model.PppoeProfile, error) {
	router, err := s.getRouter(routerID)
	if err != nil {
		return nil, err
	}
	return s.mikrotikPort.ListProfiles(router)
}

func (s *service) UpdateProfile(routerID uint, profile model.PppoeProfile) error {
	router, err := s.getRouter(routerID)
	if err != nil {
		return err
	}
	return s.mikrotikPort.UpdateProfile(router, &profile)
}

func (s *service) DeleteProfile(routerID uint, id string) error {
	router, err := s.getRouter(routerID)
	if err != nil {
		return err
	}
	return s.mikrotikPort.DeleteProfile(router, id)
}

// Session Management

func (s *service) ListActiveSessions(routerID uint) ([]model.PppoeActive, error) {
	router, err := s.getRouter(routerID)
	if err != nil {
		return nil, err
	}
	return s.mikrotikPort.ListActiveSessions(router)
}

func (s *service) ListInactiveSessions(routerID uint) ([]model.PppoeSecret, error) {
	router, err := s.getRouter(routerID)
	if err != nil {
		return nil, err
	}

	// 1. Fetch all secrets (all potential users)
	secrets, err := s.mikrotikPort.ListSecrets(router)
	if err != nil {
		return nil, err
	}

	// 2. Fetch active sessions
	activeSessions, err := s.mikrotikPort.ListActiveSessions(router)
	if err != nil {
		return nil, err
	}

	// 3. Create a map of active users for O(1) lookup
	// Key: name (username)
	activeUsers := make(map[string]bool)
	for _, session := range activeSessions {
		activeUsers[session.Name] = true
	}

	// 4. Filter secrets that are NOT in active sessions
	var inactiveUsers []model.PppoeSecret
	for _, secret := range secrets {
		if _, isActive := activeUsers[secret.Name]; !isActive && !secret.Disabled {
			inactiveUsers = append(inactiveUsers, secret)
		}
	}

	return inactiveUsers, nil
}

func (s *service) GetActiveSession(routerID uint, id string) (*model.PppoeActive, error) {
	router, err := s.getRouter(routerID)
	if err != nil {
		return nil, err
	}
	return s.mikrotikPort.GetActiveSession(router, id)
}

func (s *service) RemoveActiveSession(routerID uint, id string) error {
	router, err := s.getRouter(routerID)
	if err != nil {
		return err
	}
	return s.mikrotikPort.RemoveActiveSession(router, id)
}

// Webhooks

func (s *service) HandleOnUp(data model.PppoeCallbackData) error {
	// Construct message for realtime
	event := model.WebSocketMessage{
		Event: "session_up",
		Data:  data,
	}

	return s.cachePort.PppoePubSub().PublishSessionEvent(event.Event, event.Data)
}

func (s *service) HandleOnDown(data model.PppoeCallbackData) error {
	event := model.WebSocketMessage{
		Event: "session_down",
		Data:  data,
	}
	return s.cachePort.PppoePubSub().PublishSessionEvent(event.Event, event.Data)
}

// Realtime

func (s *service) SubscribeToEvents(ctx context.Context) (<-chan model.WebSocketMessage, error) {
	return s.cachePort.PppoePubSub().SubscribeToSessionEvents(ctx)
}
