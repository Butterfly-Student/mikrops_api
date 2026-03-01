package outbound_port

//go:generate mockgen -source=mikrotik_sync.go -destination=./../../../tests/mocks/port/mock_mikrotik_sync.go

import (
	"go-template/internal/model"
)

// MikrotikSyncMessagePort defines the interface for publishing MikroTik sync messages
type MikrotikSyncMessagePort interface {
	// PublishSyncMessage publishes a sync message to the queue for background processing
	PublishSyncMessage(message model.MikrotikSyncMessage) error
}
