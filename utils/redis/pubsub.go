package redis

import (
	"context"
)

func Publish(ctx context.Context, channel string, message string) error {
	return Client.Publish(ctx, channel, message).Err()
}

func Subscribe(ctx context.Context, channel string, handler func(string)) error {
	pubsub := Client.Subscribe(ctx, channel)
	defer pubsub.Close()

	ch := pubsub.Channel()
	for {
		select {
		case msg := <-ch:
			handler(msg.Payload)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
