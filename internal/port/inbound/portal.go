package inbound_port

type PortalHttpPort interface {
	Dashboard(a any) error
	GetProfile(a any) error
	UpdateProfile(a any) error
	UpdatePassword(a any) error
	GetSubscription(a any) error
	GetConnectionStatus(a any) error
	GetBandwidth(a any) error
	ListInvoices(a any) error
	GetInvoice(a any) error
	ListPayments(a any) error
	CreatePayment(a any) error
	GetPayment(a any) error
	UploadProof(a any) error
	ListPackages(a any) error
	ListPaymentMethods(a any) error
}
