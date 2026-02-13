package queue

import (
	"context"

	"go-template/internal/model"
)

type QueueDomain interface {
	// CRUD
	CreateQueue(routerID uint, queue model.PppoeQueue) error
	GetQueue(routerID uint, id string) (*model.PppoeQueue, error)
	ListQueues(routerID uint) ([]model.PppoeQueue, error)
	UpdateQueue(routerID uint, queue model.PppoeQueue) error
	DeleteQueue(routerID uint, id string) error

	// Realtime
	StartStreaming(routerID uint) error
	SubscribeToStats(ctx context.Context) (<-chan model.WebSocketMessage, error)
}
