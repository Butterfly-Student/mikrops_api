package outbound_port

import (
	"go-template/internal/model"
)

// PppoeCachePort defines methods for Redis operations related to PPPoE
import "context"

type PppoeCachePort interface {
	// PubSub
	PublishSessionEvent(event string, data interface{}) error
	SubscribeToSessionEvents(ctx context.Context) (<-chan model.WebSocketMessage, error)
}
