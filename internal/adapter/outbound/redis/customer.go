package redis_outbound_adapter

import (
	"context"
	"encoding/json"
	"fmt"

	"mikrops/internal/model"
	outbound_port "mikrops/internal/port/outbound"
	"mikrops/utils/redis"
)

type customerAdapter struct{}

func NewCustomerAdapter() outbound_port.CustomerCachePort {
	return &customerAdapter{}
}

func (adapter *customerAdapter) Set(data model.Customer) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("customer:%s", data.ID)
	return redis.Set(context.Background(), key, string(bytes))
}

func (adapter *customerAdapter) Get(customerID string) (model.Customer, error) {
	var customer model.Customer
	key := fmt.Sprintf("customer:%s", customerID)
	result, err := redis.Get(context.Background(), key)
	if err != nil {
		return model.Customer{}, err
	}

	err = json.Unmarshal([]byte(result), &customer)
	if err != nil {
		return model.Customer{}, err
	}

	return customer, nil
}

func (adapter *customerAdapter) Delete(customerID string) error {
	key := fmt.Sprintf("customer:%s", customerID)
	return redis.Del(context.Background(), key)
}
