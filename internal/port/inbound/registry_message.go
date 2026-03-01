package inbound_port

import (
	outbound_port "go-template/internal/port/outbound"
)

// MessagePort defines the interface for message handling
type MessagePort interface {
	Client() ClientMessagePort
	MikrotikSync() MikrotikSyncMessagePort
}

// MessagePortFactory creates MessagePort instances
type MessagePortFactory func(domain interface{}, mikrotikPort outbound_port.MikrotikPort, hotspotPort outbound_port.HotspotPort, dbPort outbound_port.DatabasePort) MessagePort
