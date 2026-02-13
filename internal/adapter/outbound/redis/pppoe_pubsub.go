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
	return a.subscribe(ctx, "pppoe-events")
}

func (a *pppoePubSubAdapter) PublishQueueStats(data interface{}) error {
	msg := model.WebSocketMessage{
		Event: "queue_stats",
		Data:  data,
	}
	bytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return redis.Publish(context.Background(), "queue-stats", string(bytes))
}

func (a *pppoePubSubAdapter) SubscribeToQueueStats(ctx context.Context) (<-chan model.WebSocketMessage, error) {
	return a.subscribe(ctx, "queue-stats")
}

func (a *pppoePubSubAdapter) subscribe(ctx context.Context, channel string) (<-chan model.WebSocketMessage, error) {
	ch := make(chan model.WebSocketMessage, 100)

	go func() {
		err := redis.Subscribe(ctx, channel, func(payload string) {
			var msg model.WebSocketMessage
			if err := json.Unmarshal([]byte(payload), &msg); err == nil {
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
