package gin_inbound_adapter

import (
	"go-template/internal/domain"
	inbound_port "go-template/internal/port/inbound"
)

type hotspotAdapter struct {
	domain domain.Domain
}

func NewHotspotAdapter(domain domain.Domain) inbound_port.HotspotHttpPort {
	return &hotspotAdapter{domain: domain}
}
