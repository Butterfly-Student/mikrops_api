package inbound_port

type HttpPort interface {
	Middleware() MiddlewareHttpPort
	Ping() PingHttpPort
	Client() ClientHttpPort
	Auth() AuthHttpPort
	User() UserHttpPort
	Pppoe() PppoeHttpPort
	Queue() QueueHttpPort
	Interface() InterfaceHttpPort
	IpPool() IpPoolHttpPort
	BandwidthProfile() BandwidthProfileHttpPort
	Customer() CustomerHttpPort
	SystemSetting() SystemSettingHttpPort
	Billing() BillingHttpPort
}
