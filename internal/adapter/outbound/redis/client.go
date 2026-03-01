package redis_outbound_adapter

import (
	"context"
	"encoding/json"

	_ "github.com/doug-martin/goqu/v9/dialect/postgres"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
	"go-template/utils/redis"
)

type clientAdapter struct{}

func NewClientAdapter() outbound_port.ClientCachePort {
	return &clientAdapter{}
}

func (adapter *clientAdapter) Set(data model.AdminUser) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	// Use email as cache key instead of BearerKey (which doesn't exist in AdminUser)
	return redis.Set(context.Background(), data.Email, string(bytes))
}

func (adapter *clientAdapter) Get(email string) (model.AdminUser, error) {
	var user model.AdminUser
	result, err := redis.Get(context.Background(), email)
	if err != nil {
		return model.AdminUser{}, err
	}

	err = json.Unmarshal([]byte(result), &user)
	if err != nil {
		return model.AdminUser{}, err
	}

	return user, nil
}
