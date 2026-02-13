package redis_outbound_adapter

import (
	"context"
	"encoding/json"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
	"go-template/utils/redis"
)

type pppoePubSubAdapter struct{}

func NewPppoePubSubAdapter() outbound_port.PppoeCachePort {
	return &pppoePubSubAdapter{}
}

func (a *pppoePubSubAdapter) PublishSessionEvent(event string, data interface{}) error {
	msg := model.WebSocketMessage{
		Event: event,
		Data:  data,
	}
	bytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return redis.Publish(context.Background(), "pppoe-events", string(bytes))
}

func (a *pppoePubSubAdapter) SubscribeToSessionEvents(ctx context.Context) (<-chan model.WebSocketMessage, error) {
	ch := make(chan model.WebSocketMessage, 100) // Buffer for safety

	go func() {
		// Subscribe blocks, so we run it in a goroutine
		// Using the provided context allows cancellation when the client disconnects
		err := redis.Subscribe(ctx, "pppoe-events", func(payload string) {
			var msg model.WebSocketMessage
			if err := json.Unmarshal([]byte(payload), &msg); err == nil {
				// Non-blocking send or drop if full?
				// For now blocking send but channel is buffered
				select {
				case ch <- msg:
				case <-ctx.Done():
					return
				}
			}
		})
		if err != nil {
			close(ch)
		}
	}()

	return ch, nil
}
