package domain

import (
	"mikrops/internal/domain/auth"
	"mikrops/internal/domain/client"
	"mikrops/internal/domain/customer"
	"mikrops/internal/domain/internet_package"
	"mikrops/internal/domain/invoice"
	"mikrops/internal/domain/mikrotik"
	"mikrops/internal/domain/nas"
	"mikrops/internal/domain/payment"
	"mikrops/internal/domain/payment_method"
	"mikrops/internal/domain/staff"
	"mikrops/internal/domain/subscription"
	"mikrops/internal/domain/tenant"
	outbound_port "mikrops/internal/port/outbound"
)

type Domain interface {
	Client() client.ClientDomain
	Tenant() tenant.TenantDomain
	Auth() auth.AuthDomain
	Staff() staff.StaffDomain
	Nas() nas.NasDomain
	InternetPackage() internet_package.InternetPackageDomain
	Customer() customer.CustomerDomain
	Subscription() subscription.SubscriptionDomain
	PaymentMethod() payment_method.PaymentMethodDomain
	Invoice() invoice.InvoiceDomain
	Payment() payment.PaymentDomain
	Mikrotik() mikrotik.MikrotikDomain
}

type domain struct {
	databasePort outbound_port.DatabasePort
	messagePort  outbound_port.MessagePort
	cachePort    outbound_port.CachePort
	workflowPort outbound_port.WorkflowPort
	httpPort     outbound_port.HttpPort
}

func NewDomain(
	databasePort outbound_port.DatabasePort,
	messagePort outbound_port.MessagePort,
	cachePort outbound_port.CachePort,
	workflowPort outbound_port.WorkflowPort,
	httpPort outbound_port.HttpPort,
) Domain {
	return &domain{
		databasePort: databasePort,
		messagePort:  messagePort,
		cachePort:    cachePort,
		workflowPort: workflowPort,
		httpPort:     httpPort,
	}
}

func (d *domain) Client() client.ClientDomain {
	return client.NewClientDomain(d.databasePort, d.messagePort, d.cachePort, d.workflowPort)
}

func (d *domain) Tenant() tenant.TenantDomain {
	return tenant.NewTenantDomain(d.databasePort, d.messagePort, d.cachePort, d.workflowPort)
}

func (d *domain) Auth() auth.AuthDomain {
	return auth.NewAuthDomain(d.databasePort, d.messagePort, d.cachePort, d.workflowPort)
}

func (d *domain) Staff() staff.StaffDomain {
	return staff.NewStaffDomain(d.databasePort, d.messagePort, d.cachePort, d.workflowPort)
}

func (d *domain) Nas() nas.NasDomain {
	return nas.NewNasDomain(d.databasePort, d.messagePort, d.cachePort, d.workflowPort, d.httpPort)
}

func (d *domain) InternetPackage() internet_package.InternetPackageDomain {
	return internet_package.NewInternetPackageDomain(d.databasePort, d.messagePort, d.cachePort, d.workflowPort)
}

func (d *domain) Customer() customer.CustomerDomain {
	return customer.NewCustomerDomain(d.databasePort, d.messagePort, d.cachePort, d.workflowPort)
}

func (d *domain) Subscription() subscription.SubscriptionDomain {
	return subscription.NewSubscriptionDomain(d.databasePort, d.messagePort, d.cachePort, d.workflowPort, d.httpPort)
}

func (d *domain) PaymentMethod() payment_method.PaymentMethodDomain {
	return payment_method.NewPaymentMethodDomain(d.databasePort, d.messagePort, d.cachePort, d.workflowPort)
}

func (d *domain) Invoice() invoice.InvoiceDomain {
	return invoice.NewInvoiceDomain(d.databasePort, d.messagePort, d.cachePort, d.workflowPort)
}

func (d *domain) Payment() payment.PaymentDomain {
	return payment.NewPaymentDomain(d.databasePort, d.messagePort, d.cachePort, d.workflowPort)
}

func (d *domain) Mikrotik() mikrotik.MikrotikDomain {
	return mikrotik.NewMikrotikDomain(d.databasePort, d.messagePort, d.cachePort, d.workflowPort, d.httpPort)
}
