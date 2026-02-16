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
	return NewBandwidthProfileHandler(s.domain)
}

func (s *adapter) Customer() inbound_port.CustomerHttpPort {
	return NewCustomerHandler(s.domain)
}

func (s *adapter) SystemSetting() inbound_port.SystemSettingHttpPort {
	return NewSystemSettingHandler(s.domain)
}

func (s *adapter) Billing() inbound_port.BillingHttpPort {
	return NewBillingHandler(s.domain)
}

func (s *adapter) Payment() inbound_port.PaymentHttpPort {
	return NewPaymentAdapter(s.domain)
}

func (s *adapter) Cash() inbound_port.CashHttpPort {
	return NewCashHandler(s.domain)
}

func (s *adapter) Notification() inbound_port.NotificationHttpPort {
	return NewNotificationHandler(s.domain)
}

func (s *adapter) Activity() inbound_port.ActivityHttpPort {
	return NewActivityHandler(s.domain)
}

func (s *adapter) Refund() inbound_port.RefundHttpPort {
	return NewRefundHttpHandler(s.domain)
}
