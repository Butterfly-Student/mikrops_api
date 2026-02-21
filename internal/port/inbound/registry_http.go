package inbound_port

type HttpPort interface {
	Payment() PaymentHttpPort
	Invoice() InvoiceHttpPort
	Customer() CustomerHttpPort
	BandwidthProfile() BandwidthProfileHttpPort
	Middleware() MiddlewareHttpPort
	Ping() PingHttpPort
	Client() ClientHttpPort
	Auth() AuthHttpPort
	User() UserHttpPort
	Pppoe() PppoeHttpPort
	Queue() QueueHttpPort
	Interface() InterfaceHttpPort
	IpPool() IpPoolHttpPort
	MikrotikRouter() MikrotikRouterHttpPort
	Hotspot() HotspotHttpPort
	Registration() RegistrationHttpPort
	CustomerPortal() CustomerPortalHttpPort
}
