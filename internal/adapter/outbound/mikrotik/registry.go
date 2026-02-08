package mikrotik_outbound_adapter

import (
	outbound_port "mikrops/internal/port/outbound"
)

type Registry struct{}

func NewRegistry() outbound_port.MikrotikPort {
	return NewAdapter()
}
