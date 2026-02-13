package queue

import (
	"context"
	"log"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
)

type service struct {
	dbPort       outbound_port.DatabasePort
	cachePort    outbound_port.CachePort
	mikrotikPort outbound_port.MikrotikPort
}

func NewQueueDomain(
	dbPort outbound_port.DatabasePort,
	cachePort outbound_port.CachePort,
	mikrotikPort outbound_port.MikrotikPort,
) QueueDomain {
	return &service{
		dbPort:       dbPort,
		cachePort:    cachePort,
		mikrotikPort: mikrotikPort,
	}
}

func (s *service) getRouter(routerID uint) (*model.MikrotikRouter, error) {
	return s.dbPort.Mikrotik().FindByID(routerID)
}

// CRUD

func (s *service) CreateQueue(routerID uint, queue model.PppoeQueue) error {
	router, err := s.getRouter(routerID)
	if err != nil {
		return err
	}
	return s.mikrotikPort.CreateQueue(router, &queue)
}

func (s *service) GetQueue(routerID uint, id string) (*model.PppoeQueue, error) {
	router, err := s.getRouter(routerID)
	if err != nil {
		return nil, err
	}
	return s.mikrotikPort.GetQueue(router, id)
}

func (s *service) ListQueues(routerID uint) ([]model.PppoeQueue, error) {
	router, err := s.getRouter(routerID)
	if err != nil {
		return nil, err
	}
	return s.mikrotikPort.ListQueues(router)
}

func (s *service) UpdateQueue(routerID uint, queue model.PppoeQueue) error {
	router, err := s.getRouter(routerID)
	if err != nil {
		return err
	}
	return s.mikrotikPort.UpdateQueue(router, &queue)
}

func (s *service) DeleteQueue(routerID uint, id string) error {
	router, err := s.getRouter(routerID)
	if err != nil {
		return err
	}
	return s.mikrotikPort.DeleteQueue(router, id)
}

// Streaming

func (s *service) StartStreaming(routerID uint) error {
	router, err := s.getRouter(routerID)
	if err != nil {
		return err
	}

	statsCh, err := s.mikrotikPort.ListenQueueStats(router)
	if err != nil {
		return err
	}

	go func() {
		for stats := range statsCh {
			// Publish to Redis
			msg := model.WebSocketMessage{
				Event: "queue_stats",
				Data:  stats,
			}
			// Use generic PublishQueueStats or SessionEvent with specialized logic?
			// Using specific method on CachePort for clarity
			if err := s.cachePort.PppoePubSub().PublishQueueStats(msg); err != nil {
				log.Printf("Failed to publish queue stats: %v", err)
			}
		}
	}()

	return nil
}

func (s *service) SubscribeToStats(ctx context.Context) (<-chan model.WebSocketMessage, error) {
	return s.cachePort.PppoePubSub().SubscribeToQueueStats(ctx)
}
