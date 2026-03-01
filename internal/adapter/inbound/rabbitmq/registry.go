package rabbitmq_inbound_adapter

import (
	"go-template/internal/domain"
	inbound_port "go-template/internal/port/inbound"
	outbound_port "go-template/internal/port/outbound"
)

type adapter struct {
	domain       domain.Domain
	mikrotikPort outbound_port.MikrotikPort
	hotspotPort  outbound_port.HotspotPort
	dbPort       outbound_port.DatabasePort
}

func NewAdapter(
	domain domain.Domain,
	mikrotikPort outbound_port.MikrotikPort,
	hotspotPort outbound_port.HotspotPort,
	dbPort outbound_port.DatabasePort,
) inbound_port.MessagePort {
	return &adapter{
		domain:       domain,
		mikrotikPort: mikrotikPort,
		hotspotPort:  hotspotPort,
		dbPort:       dbPort,
	}
}

func (a *adapter) Client() inbound_port.ClientMessagePort {
	return NewClientAdapter(a.domain)
}

func (a *adapter) MikrotikSync() inbound_port.MikrotikSyncMessagePort {
	return NewMikrotikSyncAdapter(a.mikrotikPort, a.hotspotPort, a.dbPort)
}
