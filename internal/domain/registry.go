package domain

import (
	"mikrops/internal/domain/activity_log"
	"mikrops/internal/domain/auth"
	"mikrops/internal/domain/client"
	"mikrops/internal/domain/customer"
	"mikrops/internal/domain/customer_registration"
	"mikrops/internal/domain/cutoff"
	"mikrops/internal/domain/internet_package"
	"mikrops/internal/domain/invoice"
	"mikrops/internal/domain/mikrotik"
	"mikrops/internal/domain/mikrotik_sync_log"
	"mikrops/internal/domain/nas"
	"mikrops/internal/domain/payment"
	"mikrops/internal/domain/payment_method"
	"mikrops/internal/domain/pppoe_account"
	"mikrops/internal/domain/staff"
	"mikrops/internal/domain/subscription"
	"mikrops/internal/domain/tenant"
	"mikrops/internal/domain/tenant_setting"
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
	TenantSetting() tenant_setting.TenantSettingDomain
	CustomerRegistration() customer_registration.CustomerRegistrationDomain
	PppoeAccount() pppoe_account.PppoeAccountDomain
	ActivityLog() activity_log.ActivityLogDomain
	MikrotikSyncLog() mikrotik_sync_log.MikrotikSyncLogDomain
	Cutoff() cutoff.CutoffDomain
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
	return invoice.NewInvoiceDomain(d.databasePort, d.messagePort, d.cachePort, d.workflowPort, d.httpPort)
}

func (d *domain) Payment() payment.PaymentDomain {
	return payment.NewPaymentDomain(d.databasePort, d.messagePort, d.cachePort, d.workflowPort)
}

func (d *domain) Mikrotik() mikrotik.MikrotikDomain {
	return mikrotik.NewMikrotikDomain(d.databasePort, d.messagePort, d.cachePort, d.workflowPort, d.httpPort)
}

func (d *domain) TenantSetting() tenant_setting.TenantSettingDomain {
	return tenant_setting.NewTenantSettingDomain(d.databasePort, d.messagePort, d.cachePort, d.workflowPort)
}

func (d *domain) CustomerRegistration() customer_registration.CustomerRegistrationDomain {
	return customer_registration.NewCustomerRegistrationDomain(d.databasePort, d.messagePort, d.cachePort, d.workflowPort)
}

func (d *domain) PppoeAccount() pppoe_account.PppoeAccountDomain {
	return pppoe_account.NewPppoeAccountDomain(d.databasePort, d.messagePort, d.cachePort, d.workflowPort)
}

func (d *domain) ActivityLog() activity_log.ActivityLogDomain {
	return activity_log.NewActivityLogDomain(d.databasePort, d.messagePort, d.cachePort, d.workflowPort)
}

func (d *domain) MikrotikSyncLog() mikrotik_sync_log.MikrotikSyncLogDomain {
	return mikrotik_sync_log.NewMikrotikSyncLogDomain(d.databasePort, d.messagePort, d.cachePort, d.workflowPort)
}

func (d *domain) Cutoff() cutoff.CutoffDomain {
	return cutoff.NewCutoffDomain(d.databasePort, d.messagePort, d.cachePort, d.workflowPort)
}
