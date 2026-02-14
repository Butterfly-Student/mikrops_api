package queue

import (
	"context"
	"log"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
)

type QueueDomain interface {
	// CRUD
	CreateQueue(routerID string, queue model.PppoeQueue) error
	GetQueue(routerID string, id string) (*model.PppoeQueue, error)
	ListQueues(routerID string) ([]model.PppoeQueue, error)
	UpdateQueue(routerID string, queue model.PppoeQueue) error
	DeleteQueue(routerID string, id string) error

	// Realtime
	StartStreaming(routerID string) error
	SubscribeToStats(ctx context.Context) (<-chan model.WebSocketMessage, error)

	// Enhanced streaming with context and filtering
	StartStreamingAll(ctx context.Context, routerID string) error
	StartStreamingByName(ctx context.Context, routerID string, queueName string) error
	StopStreamingAll(routerID string) error
	StopStreamingByName(routerID string, queueName string) error
	SubscribeToStatsByName(ctx context.Context, queueName string) (<-chan model.WebSocketMessage, error)
}

type domain struct {
	dbPort       outbound_port.DatabasePort
	cachePort    outbound_port.CachePort
	mikrotikPort outbound_port.MikrotikPort
}

func NewQueueDomain(
	dbPort outbound_port.DatabasePort,
	cachePort outbound_port.CachePort,
	mikrotikPort outbound_port.MikrotikPort,
) QueueDomain {
	return &domain{
		dbPort:       dbPort,
		cachePort:    cachePort,
		mikrotikPort: mikrotikPort,
	}
}

func (d *domain) getRouter(routerID string) (*model.MikrotikRouter, error) {
	return d.dbPort.Mikrotik().FindByID(routerID)
}

// CRUD

func (d *domain) CreateQueue(routerID string, queue model.PppoeQueue) error {
	router, err := d.getRouter(routerID)
	if err != nil {
		return err
	}
	return d.mikrotikPort.CreateQueue(router, &queue)
}

func (d *domain) GetQueue(routerID string, id string) (*model.PppoeQueue, error) {
	router, err := d.getRouter(routerID)
	if err != nil {
		return nil, err
	}
	return d.mikrotikPort.GetQueue(router, id)
}

func (d *domain) ListQueues(routerID string) ([]model.PppoeQueue, error) {
	router, err := d.getRouter(routerID)
	if err != nil {
		return nil, err
	}
	return d.mikrotikPort.ListQueues(router)
}

func (d *domain) UpdateQueue(routerID string, queue model.PppoeQueue) error {
	router, err := d.getRouter(routerID)
	if err != nil {
		return err
	}
	return d.mikrotikPort.UpdateQueue(router, &queue)
}

func (d *domain) DeleteQueue(routerID string, id string) error {
	router, err := d.getRouter(routerID)
	if err != nil {
		return err
	}
	return d.mikrotikPort.DeleteQueue(router, id)
}

// Streaming

func (d *domain) StartStreaming(routerID string) error {
	router, err := d.getRouter(routerID)
	if err != nil {
		return err
	}

	statsCh, err := d.mikrotikPort.ListenQueueStats(router)
	if err != nil {
		return err
	}

	go func() {
		for stats := range statsCh {
			msg := model.WebSocketMessage{
				Event: "queue_stats",
				Data:  stats,
			}
			if err := d.cachePort.PppoePubSub().PublishQueueStats(msg); err != nil {
				log.Printf("Failed to publish queue stats: %v", err)
			}
		}
	}()

	return nil
}

func (d *domain) SubscribeToStats(ctx context.Context) (<-chan model.WebSocketMessage, error) {
	return d.cachePort.PppoePubSub().SubscribeToQueueStats(ctx)
}

// Enhanced streaming methods with context and filtering

func (d *domain) StartStreamingAll(ctx context.Context, routerID string) error {
	router, err := d.getRouter(routerID)
	if err != nil {
		return err
	}

	statsCh, err := d.mikrotikPort.ListenQueueStatsWithContext(ctx, router)
	if err != nil {
		return err
	}

	go func() {
		for stats := range statsCh {
			msg := model.WebSocketMessage{
				Event: "queue_stats",
				Data:  stats,
			}
			if err := d.cachePort.PppoePubSub().PublishQueueStats(msg); err != nil {
				log.Printf("Failed to publish queue stats: %v", err)
			}
		}
	}()

	return nil
}

func (d *domain) StartStreamingByName(ctx context.Context, routerID string, queueName string) error {
	router, err := d.getRouter(routerID)
	if err != nil {
		return err
	}

	statsCh, err := d.mikrotikPort.ListenQueueStatsByName(ctx, router, queueName)
	if err != nil {
		return err
	}

	go func() {
		for stats := range statsCh {
			msg := model.WebSocketMessage{
				Event: "queue_stats",
				Data:  stats,
			}
			if err := d.cachePort.PppoePubSub().PublishQueueStatsByName(queueName, msg); err != nil {
				log.Printf("Failed to publish queue stats for %s: %v", queueName, err)
			}
		}
	}()

	return nil
}

func (d *domain) StopStreamingAll(routerID string) error {
	return nil
}

func (d *domain) StopStreamingByName(routerID string, queueName string) error {
	return nil
}

func (d *domain) SubscribeToStatsByName(ctx context.Context, queueName string) (<-chan model.WebSocketMessage, error) {
	return d.cachePort.PppoePubSub().SubscribeToQueueStatsByName(ctx, queueName)
}
