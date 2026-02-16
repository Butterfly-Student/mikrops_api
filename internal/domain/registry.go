package domain

import (
	"go-template/internal/domain/activity"
	"go-template/internal/domain/auth"
	"go-template/internal/domain/bandwidth_profile"
	"go-template/internal/domain/billing"
	"go-template/internal/domain/cash"
	"go-template/internal/domain/client"
	"go-template/internal/domain/customer"
	"go-template/internal/domain/iface"
	"go-template/internal/domain/ippool"
	"go-template/internal/domain/mikrotik"
	"go-template/internal/domain/notification"
	"go-template/internal/domain/payment"
	"go-template/internal/domain/ping"
	"go-template/internal/domain/pppoe"
	"go-template/internal/domain/queue"
	"go-template/internal/domain/refund"
	"go-template/internal/domain/system_setting"
	"go-template/internal/domain/user"
	outbound_port "go-template/internal/port/outbound"
	"go-template/utils/email"
	"go-template/utils/gowa"
	"go-template/utils/xendit"

	"github.com/casbin/casbin/v3"
)

type Domain interface {
	Client() client.ClientDomain
	Auth() auth.AuthDomain
	User() user.UserDomain
	Mikrotik() mikrotik.MikrotikDomain
	BandwidthProfile() bandwidth_profile.BandwidthProfileDomain
	Customer() customer.CustomerDomain
	SystemSetting() system_setting.SystemSettingDomain
	Pppoe() pppoe.PppoeDomain
	Queue() queue.QueueDomain
	Interface() iface.InterfaceDomain
	IpPool() ippool.IpPoolDomain
	Ping() ping.PingDomain
	Billing() billing.BillingDomain
	Payment() payment.PaymentDomain
	Cash() cash.CashDomain
	Refund() refund.RefundDomain
	Activity() activity.ActivityDomain
	Notification() notification.NotificationDomain
	Workflow() outbound_port.WorkflowPort
}

type domain struct {
	databasePort   outbound_port.DatabasePort
	messagePort    outbound_port.MessagePort
	cachePort      outbound_port.CachePort
	workflowPort   outbound_port.WorkflowPort
	mikrotikPort   outbound_port.MikrotikPort
	emailUtil      *email.EmailUtil
	gowaUtil       *gowa.Client
	enforcer       *casbin.Enforcer
	activityDomain activity.ActivityDomain
}

func NewDomain(
	databasePort outbound_port.DatabasePort,
	messagePort outbound_port.MessagePort,
	cachePort outbound_port.CachePort,
	workflowPort outbound_port.WorkflowPort,
	mikrotikPort outbound_port.MikrotikPort,
	emailUtil *email.EmailUtil,
	gowaUtil *gowa.Client,
	enforcer *casbin.Enforcer,
) Domain {
	return &domain{
		databasePort:   databasePort,
		messagePort:    messagePort,
		cachePort:      cachePort,
		workflowPort:   workflowPort,
		mikrotikPort:   mikrotikPort,
		emailUtil:      emailUtil,
		gowaUtil:       gowaUtil,
		enforcer:       enforcer,
		activityDomain: activity.NewActivityDomain(databasePort),
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

func (d *domain) Mikrotik() mikrotik.MikrotikDomain {
	return mikrotik.NewMikrotikDomain(d.databasePort.Mikrotik(), d.mikrotikPort)
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
	return customer.NewCustomerDomain(d.databasePort, d.mikrotikPort, d.databasePort.BandwidthProfile())
}

func (d *domain) SystemSetting() system_setting.SystemSettingDomain {
	return system_setting.NewSystemSettingDomain(d.databasePort)
}

func (d *domain) Billing() billing.BillingDomain {
	return billing.NewBillingDomain(d.databasePort, d.databasePort.SystemSetting())
}

func (d *domain) Payment() payment.PaymentDomain {
	return payment.NewPaymentDomain(d.databasePort, xendit.GetClient())
}

func (d *domain) Cash() cash.CashDomain {
	return cash.NewCashDomain(d.databasePort)
}

func (d *domain) Notification() notification.NotificationDomain {
	return notification.NewNotificationDomain(d.databasePort, d.emailUtil, d.gowaUtil)
}

func (d *domain) Activity() activity.ActivityDomain {
	return d.activityDomain
}

func (d *domain) Refund() refund.RefundDomain {
	return refund.NewRefundDomain(d.databasePort.Refund())
}

func (d *domain) Workflow() outbound_port.WorkflowPort {
	return d.workflowPort
}
