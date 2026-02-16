package redis

import (
	"context"
	"os"

	redis "github.com/redis/go-redis/v9"
)

var Client *redis.Client

func InitDatabase() {
	addr := os.Getenv("CACHE_HOST")
	port := os.Getenv("CACHE_PORT")
	pass := os.Getenv("CACHE_PASSWORD")
	if port == "" {
		port = "6379"
	}
	if addr == "" {
		addr = os.Getenv("MESSAGE_HOST")
	}
	Client = redis.NewClient(&redis.Options{
		Addr:     addr + ":" + port,
		Password: pass,
	})
}

func Set(ctx context.Context, key string, value interface{}) error {
	return Client.Set(ctx, key, value, 24*60*60*1e9).Err() // 1 day in nanoseconds
}

func Get(ctx context.Context, key string) (string, error) {
	return Client.Get(ctx, key).Result()
}

func Del(ctx context.Context, key string) error {
	return Client.Del(ctx, key).Err()
}
