package gin_inbound_adapter

import (
	"mikrops/internal/domain"
	inbound_port "mikrops/internal/port/inbound"
	outbound_port "mikrops/internal/port/outbound"
)

type adapter struct {
	domain   domain.Domain
	httpPort outbound_port.HttpPort
}

func NewAdapter(
	domain domain.Domain,
	httpPort outbound_port.HttpPort,
) inbound_port.HttpPort {
	return &adapter{
		domain:   domain,
		httpPort: httpPort,
	}
}

func (s *adapter) Ping() inbound_port.PingHttpPort {
	return NewPingAdapter()
}

func (s *adapter) Middleware() inbound_port.MiddlewareHttpPort {
	return NewMiddlewareAdapter(s.domain)
}

func (s *adapter) Client() inbound_port.ClientHttpPort {
	return NewClientAdapter(s.domain)
}

func (s *adapter) Auth() inbound_port.AuthHttpPort {
	return NewAuthAdapter(s.domain)
}

func (s *adapter) Tenant() inbound_port.TenantHttpPort {
	return NewTenantAdapter(s.domain)
}

func (s *adapter) Staff() inbound_port.StaffHttpPort {
	return NewStaffAdapter(s.domain)
}

func (s *adapter) Nas() inbound_port.NasHttpPort {
	return NewNasAdapter(s.domain)
}

func (s *adapter) InternetPackage() inbound_port.InternetPackageHttpPort {
	return NewInternetPackageAdapter(s.domain)
}

func (s *adapter) Customer() inbound_port.CustomerHttpPort {
	return NewCustomerAdapter(s.domain)
}

func (s *adapter) Subscription() inbound_port.SubscriptionHttpPort {
	return NewSubscriptionAdapter(s.domain)
}

func (s *adapter) PaymentMethod() inbound_port.PaymentMethodHttpPort {
	return NewPaymentMethodAdapter(s.domain)
}

func (s *adapter) Invoice() inbound_port.InvoiceHttpPort {
	return NewInvoiceAdapter(s.domain)
}

func (s *adapter) Payment() inbound_port.PaymentHttpPort {
	return NewPaymentAdapter(s.domain)
}

func (s *adapter) Portal() inbound_port.PortalHttpPort {
	return NewPortalAdapter(s.domain)
}

func (s *adapter) Mikrotik() inbound_port.MikrotikHttpPort {
	return NewMikrotikAdapter(s.domain, s.httpPort)
}
