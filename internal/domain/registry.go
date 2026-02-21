package domain

import (
	"go-template/internal/domain/auth"
	"go-template/internal/domain/bandwidth_profile"
	"go-template/internal/domain/client"
	"go-template/internal/domain/customer"
	customer_portal "go-template/internal/domain/customer_portal"
	hotspot_domain "go-template/internal/domain/hotspot"
	"go-template/internal/domain/iface"
	"go-template/internal/domain/invoice"
	"go-template/internal/domain/ippool"
	"go-template/internal/domain/mikrotik_router"
	"go-template/internal/domain/payment"
	"go-template/internal/domain/ping"
	"go-template/internal/domain/pppoe"
	"go-template/internal/domain/queue"
	"go-template/internal/domain/registration"
	"go-template/internal/domain/user"
	outbound_port "go-template/internal/port/outbound"

	"github.com/casbin/casbin/v3"
)

type Domain interface {
	Payment() payment.PaymentDomain
	Invoice() invoice.InvoiceDomain
	Customer() customer.CustomerDomain
	BandwidthProfile() bandwidth_profile.BandwidthProfileDomain
	Client() client.ClientDomain
	Auth() auth.AuthDomain
	User() user.UserDomain
	Pppoe() pppoe.PppoeDomain
	Queue() queue.QueueDomain
	Interface() iface.InterfaceDomain
	IpPool() ippool.IpPoolDomain
	Ping() ping.PingDomain
	MikrotikRouter() mikrotik_router.MikrotikRouterDomain
	Hotspot() hotspot_domain.HotspotDomain
	Registration() registration.RegistrationDomain
	CustomerPortal() customer_portal.CustomerPortalDomain
}

type domain struct {
	databasePort outbound_port.DatabasePort
	messagePort  outbound_port.MessagePort
	cachePort    outbound_port.CachePort
	workflowPort outbound_port.WorkflowPort
	mikrotikPort outbound_port.MikrotikPort
	hotspotPort  outbound_port.HotspotPort
	enforcer     *casbin.Enforcer
}

func NewDomain(
	databasePort outbound_port.DatabasePort,
	messagePort outbound_port.MessagePort,
	cachePort outbound_port.CachePort,
	workflowPort outbound_port.WorkflowPort,
	mikrotikPort outbound_port.MikrotikPort,
	hotspotPort outbound_port.HotspotPort,
	enforcer *casbin.Enforcer,
) Domain {
	return &domain{
		databasePort: databasePort,
		messagePort:  messagePort,
		cachePort:    cachePort,
		workflowPort: workflowPort,
		mikrotikPort: mikrotikPort,
		hotspotPort:  hotspotPort,
		enforcer:     enforcer,
	}
}

func (d *domain) Client() client.ClientDomain {
	return client.NewClientDomain(d.databasePort, d.messagePort, d.cachePort, d.workflowPort)
}

func (d *domain) Auth() auth.AuthDomain {
	return auth.NewAuthDomain(d.databasePort, d.enforcer)
}

func (d *domain) User() user.UserDomain {
	return user.NewUserDomain(d.databasePort)
}

func (d *domain) Pppoe() pppoe.PppoeDomain {
	return pppoe.NewPppoeDomain(d.databasePort, d.cachePort, d.mikrotikPort)
}

func (d *domain) Queue() queue.QueueDomain {
	return queue.NewQueueDomain(d.databasePort, d.cachePort, d.mikrotikPort)
}

func (d *domain) Interface() iface.InterfaceDomain {
	return iface.NewInterfaceDomain(d.databasePort, d.cachePort, d.mikrotikPort)
}

func (d *domain) IpPool() ippool.IpPoolDomain {
	return ippool.NewIpPoolDomain(d.databasePort, d.mikrotikPort)
}

func (d *domain) Ping() ping.PingDomain {
	return ping.NewPingDomain(d.databasePort, d.cachePort, d.mikrotikPort)
}

func (d *domain) BandwidthProfile() bandwidth_profile.BandwidthProfileDomain {
	return bandwidth_profile.NewBandwidthProfileDomain(d.databasePort, d.mikrotikPort)
}

func (d *domain) Customer() customer.CustomerDomain {
	return customer.NewCustomerDomain(d.databasePort, d.mikrotikPort)
}

func (d *domain) Invoice() invoice.InvoiceDomain {
	return invoice.NewInvoiceDomain(d.databasePort)
}

func (d *domain) Payment() payment.PaymentDomain {
	return payment.NewPaymentDomain(d.databasePort, d.Invoice(), d.Customer())
}

func (d *domain) MikrotikRouter() mikrotik_router.MikrotikRouterDomain {
	return mikrotik_router.NewMikrotikRouterDomain(d.databasePort, d.mikrotikPort)
}

func (d *domain) Hotspot() hotspot_domain.HotspotDomain {
	return hotspot_domain.NewHotspotDomain(d.databasePort, d.hotspotPort)
}

func (d *domain) Registration() registration.RegistrationDomain {
	return registration.NewRegistrationDomain(d.databasePort, d.Customer())
}

func (d *domain) CustomerPortal() customer_portal.CustomerPortalDomain {
	return customer_portal.NewCustomerPortalDomain(d.databasePort, d.Customer())
}
