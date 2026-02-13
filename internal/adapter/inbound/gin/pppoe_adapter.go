package gin_inbound_adapter

import (
	"go-template/internal/domain"
	inbound_port "go-template/internal/port/inbound"
)

type pppoeAdapter struct {
	domain domain.Domain
}

func NewPppoeAdapter(domain domain.Domain) inbound_port.PppoeHttpPort {
	return &pppoeAdapter{
		domain: domain,
	}
}
