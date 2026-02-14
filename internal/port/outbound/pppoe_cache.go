package outbound_port

import (
	"context"

	"go-template/internal/model"
)

// PppoeCachePort defines methods for Redis operations related to PPPoE
//
//go:generate mockgen -source=pppoe_cache.go -destination=./../../../tests/mocks/port/mock_pppoe_cache.go
type PppoeCachePort interface {
	// PubSub
	PublishSessionEvent(event string, data interface{}) error
	SubscribeToSessionEvents(ctx context.Context) (<-chan model.WebSocketMessage, error)

	// Queue PubSub
	PublishQueueStats(data interface{}) error
	SubscribeToQueueStats(ctx context.Context) (<-chan model.WebSocketMessage, error)

	// Queue PubSub by name
	PublishQueueStatsByName(queueName string, data interface{}) error
	SubscribeToQueueStatsByName(ctx context.Context, queueName string) (<-chan model.WebSocketMessage, error)

	// Interface PubSub
	PublishInterfaceStats(data interface{}) error
	SubscribeToInterfaceStats(ctx context.Context) (<-chan model.WebSocketMessage, error)

	// Interface PubSub by name
	PublishInterfaceStatsByName(interfaceName string, data interface{}) error
	SubscribeToInterfaceStatsByName(ctx context.Context, interfaceName string) (<-chan model.WebSocketMessage, error)

	// Ping PubSub
	PublishPingResult(address string, data interface{}) error
	SubscribeToPingResults(ctx context.Context, address string) (<-chan model.WebSocketMessage, error)
}
