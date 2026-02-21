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

func (s *adapter) Pppoe() inbound_port.PppoeHttpPort {
	return NewPppoeAdapter(s.domain)
}

func (s *adapter) Queue() inbound_port.QueueHttpPort {
	return NewQueueAdapter(s.domain)
}

func (s *adapter) Interface() inbound_port.InterfaceHttpPort {
	return NewInterfaceAdapter(s.domain)
}

func (s *adapter) IpPool() inbound_port.IpPoolHttpPort {
	return NewIpPoolAdapter(s.domain)
}

func (s *adapter) BandwidthProfile() inbound_port.BandwidthProfileHttpPort {
	return NewBandwidthProfileAdapter(s.domain)
}

func (s *adapter) Customer() inbound_port.CustomerHttpPort {
	return NewCustomerAdapter(s.domain)
}

func (s *adapter) Invoice() inbound_port.InvoiceHttpPort {
	return NewInvoiceAdapter(s.domain)
}

func (s *adapter) Payment() inbound_port.PaymentHttpPort {
	return NewPaymentAdapter(s.domain)
}

func (s *adapter) MikrotikRouter() inbound_port.MikrotikRouterHttpPort {
	return NewMikrotikRouterAdapter(s.domain)
}

func (s *adapter) Hotspot() inbound_port.HotspotHttpPort {
	return NewHotspotAdapter(s.domain)
}

func (s *adapter) Registration() inbound_port.RegistrationHttpPort {
	return NewRegistrationAdapter(s.domain)
}

func (s *adapter) CustomerPortal() inbound_port.CustomerPortalHttpPort {
	return NewCustomerPortalAdapter(s.domain)
}
