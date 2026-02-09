package http_outbound_adapter

import (
	outbound_port "mikrops/internal/port/outbound"

	mikrotik_adapter "mikrops/internal/adapter/outbound/mikrotik"
	"mikrops/internal/adapter/outbound/whatsapp"
)

type adapter struct {
	mikrotikClient outbound_port.MikrotikPort
	whatsappClient outbound_port.WhatsappPort
}

func NewAdapter() outbound_port.HttpPort {
	return &adapter{
		mikrotikClient: mikrotik_adapter.NewAdapter(),
		whatsappClient: whatsapp.NewWhatsappClient(),
	}
}

func (a *adapter) Mikrotik() outbound_port.MikrotikPort {
	return a.mikrotikClient
}

func (a *adapter) Whatsapp() outbound_port.WhatsappPort {
	return a.whatsappClient
}
