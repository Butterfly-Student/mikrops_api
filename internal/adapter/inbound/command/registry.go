package command_inbound_adapter

import (
	"mikrops/internal/domain"
	inbound_port "mikrops/internal/port/inbound"
)

type adapter struct {
	domain domain.Domain
}

func NewAdapter(
	domain domain.Domain,
) inbound_port.CommandPort {
	return &adapter{
		domain: domain,
	}
}

func (s *adapter) Client() inbound_port.ClientCommandPort {
	return NewClientAdapter(s.domain)
}
