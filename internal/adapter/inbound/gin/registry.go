package gin_inbound_adapter

import (
	"go-template/internal/domain"
	inbound_port "go-template/internal/port/inbound"
)

type adapter struct {
	domain domain.Domain
}

func NewAdapter(domain domain.Domain) inbound_port.HttpPort {
	return &adapter{
		domain: domain,
	}
}

func (s *adapter) Client() inbound_port.ClientHttpPort {
	return NewClientAdapter(s.domain)
}

func (s *adapter) Middleware() inbound_port.MiddlewareHttpPort {
	return NewMiddlewareAdapter(s.domain)
}

func (s *adapter) Ping() inbound_port.PingHttpPort {
	return NewPingAdapter(s.domain)
}

func (s *adapter) Auth() inbound_port.AuthHttpPort {
	return NewAuthAdapter(s.domain)
}

func (s *adapter) User() inbound_port.UserHttpPort {
	return NewUserAdapter(s.domain)
}
