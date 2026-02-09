package http_outbound_adapter

import (
	outbound_port "mikrops/internal/port/outbound"

	mikrotik_adapter "mikrops/internal/adapter/outbound/mikrotik"
	"mikrops/internal/adapter/outbound/whatsapp"
	xendit_adapter "mikrops/internal/adapter/outbound/xendit"
)

type adapter struct {
	mikrotikClient outbound_port.MikrotikPort
	whatsappClient outbound_port.WhatsappPort
	xenditClient   outbound_port.XenditPort
}

func NewAdapter() outbound_port.HttpPort {
	return &adapter{
		mikrotikClient: mikrotik_adapter.NewAdapter(),
		whatsappClient: whatsapp.NewWhatsappClient(),
		xenditClient:   xendit_adapter.NewAdapter(),
	}
}

func (a *adapter) Mikrotik() outbound_port.MikrotikPort {
	return a.mikrotikClient
}

func (a *adapter) Whatsapp() outbound_port.WhatsappPort {
	return a.whatsappClient
}

func (a *adapter) Xendit() outbound_port.XenditPort {
	return a.xenditClient
}
