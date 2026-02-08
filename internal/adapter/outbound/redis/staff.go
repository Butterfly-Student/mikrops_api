package redis_outbound_adapter

import (
	"context"
	"encoding/json"
	"fmt"

	"mikrops/internal/model"
	outbound_port "mikrops/internal/port/outbound"
	"mikrops/utils/redis"
)

type staffAdapter struct{}

func NewStaffAdapter() outbound_port.StaffCachePort {
	return &staffAdapter{}
}

func (adapter *staffAdapter) Set(data model.Staff) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("staff:%s", data.ID)
	return redis.Set(context.Background(), key, string(bytes))
}

func (adapter *staffAdapter) Get(staffID string) (model.Staff, error) {
	var staff model.Staff
	key := fmt.Sprintf("staff:%s", staffID)
	result, err := redis.Get(context.Background(), key)
	if err != nil {
		return model.Staff{}, err
	}

	err = json.Unmarshal([]byte(result), &staff)
	if err != nil {
		return model.Staff{}, err
	}

	return staff, nil
}

func (adapter *staffAdapter) Delete(staffID string) error {
	key := fmt.Sprintf("staff:%s", staffID)
	return redis.Del(context.Background(), key)
}
