package outbound_port

import (
	"context"

	"go-template/internal/model"
)

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

	// Queue
	CreateQueue(router *model.MikrotikRouter, queue *model.PppoeQueue) error
	UpdateQueue(router *model.MikrotikRouter, queue *model.PppoeQueue) error
	DeleteQueue(router *model.MikrotikRouter, id string) error
	GetQueue(router *model.MikrotikRouter, id string) (*model.PppoeQueue, error)
	ListQueues(router *model.MikrotikRouter) ([]model.PppoeQueue, error)

	// Queue Streaming
	// ListenQueueStats starts streaming stats for all simple queues
	// It returns a channel for updates and a cancel function/channel
	ListenQueueStats(router *model.MikrotikRouter) (<-chan []model.QueueStats, error)

	// ListenQueueStatsWithContext streams all queue stats with context support for cancellation
	ListenQueueStatsWithContext(ctx context.Context, router *model.MikrotikRouter) (<-chan []model.QueueStats, error)

	// ListenQueueStatsByName streams specific queue stats by name with context support
	ListenQueueStatsByName(ctx context.Context, router *model.MikrotikRouter, queueName string) (<-chan model.QueueStats, error)

	// Interface Monitoring
	// MonitorAllInterfaces streams traffic stats for all interfaces
	MonitorAllInterfaces(ctx context.Context, router *model.MikrotikRouter) (<-chan []model.InterfaceStats, error)

	// MonitorInterface streams traffic stats for a specific interface by name
	MonitorInterface(ctx context.Context, router *model.MikrotikRouter, interfaceName string) (<-chan model.InterfaceStats, error)

	// IP Pool CRUD
	CreateIpPool(router *model.MikrotikRouter, pool *model.IpPool) error
	UpdateIpPool(router *model.MikrotikRouter, pool *model.IpPool) error
	DeleteIpPool(router *model.MikrotikRouter, id string) error
	GetIpPool(router *model.MikrotikRouter, id string) (*model.IpPool, error)
	ListIpPools(router *model.MikrotikRouter) ([]model.IpPool, error)

	// Ping
	Ping(ctx context.Context, router *model.MikrotikRouter, req model.PingRequest) (<-chan model.PingResult, error)

	// Isolation
	SetupIsolation(router *model.MikrotikRouter, config model.IsolationConfig) error
	CheckIsolationSetup(router *model.MikrotikRouter) (bool, error)
}

// MikrotikDatabasePort defines methods for managing MikroTik router configurations
type MikrotikDatabasePort interface {
	Create(router *model.MikrotikRouter) error
	FindByID(id string) (*model.MikrotikRouter, error)
	FindAll() ([]model.MikrotikRouter, error)
	Update(router *model.MikrotikRouter) error
	Delete(id string) error
}
