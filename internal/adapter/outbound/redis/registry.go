package redis_outbound_adapter

import (
	outbound_port "mikrops/internal/port/outbound"
)

type adapter struct {
}

func NewAdapter() outbound_port.CachePort {
	return &adapter{}
}

func (s *adapter) Client() outbound_port.ClientCachePort {
	return NewClientAdapter()
}

func (s *adapter) Staff() outbound_port.StaffCachePort {
	return NewStaffAdapter()
}

func (s *adapter) Customer() outbound_port.CustomerCachePort {
	return NewCustomerAdapter()
}
