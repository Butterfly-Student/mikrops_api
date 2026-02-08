package http_outbound_adapter

import (
	mikrotik_outbound_adapter "mikrops/internal/adapter/outbound/mikrotik"
	outbound_port "mikrops/internal/port/outbound"
)

type adapter struct{}

func NewAdapter() outbound_port.HttpPort {
	return &adapter{}
}

func (s *adapter) Mikrotik() outbound_port.MikrotikPort {
	return mikrotik_outbound_adapter.NewRegistry()
}
