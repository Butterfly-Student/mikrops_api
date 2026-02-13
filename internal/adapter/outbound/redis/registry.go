package redis_outbound_adapter

import (
	outbound_port "go-template/internal/port/outbound"
)

type adapter struct {
}

func NewAdapter() outbound_port.CachePort {
	return &adapter{}
}

func (s *adapter) Client() outbound_port.ClientCachePort {
	return NewClientAdapter()
}

func (s *adapter) PppoePubSub() outbound_port.PppoeCachePort {
	return NewPppoePubSubAdapter()
}
