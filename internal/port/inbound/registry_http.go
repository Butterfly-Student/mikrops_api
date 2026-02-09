package inbound_port

type HttpPort interface {
	Middleware() MiddlewareHttpPort
	Ping() PingHttpPort
	Client() ClientHttpPort
	Auth() AuthHttpPort
	Tenant() TenantHttpPort
	Staff() StaffHttpPort
	Nas() NasHttpPort
	InternetPackage() InternetPackageHttpPort
	Customer() CustomerHttpPort
	Subscription() SubscriptionHttpPort
	PaymentMethod() PaymentMethodHttpPort
	Invoice() InvoiceHttpPort
	Payment() PaymentHttpPort
	Portal() PortalHttpPort
	Mikrotik() MikrotikHttpPort
	TenantSetting() TenantSettingHttpPort
	CustomerRegistration() CustomerRegistrationHttpPort
	PppoeAccount() PppoeAccountHttpPort
	ActivityLog() ActivityLogHttpPort
	MikrotikSyncLog() MikrotikSyncLogHttpPort
}
