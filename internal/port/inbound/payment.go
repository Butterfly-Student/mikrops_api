package inbound_port

type PaymentHttpPort interface {
	List(a any) error
	Get(a any) error
	Verify(a any) error
	Reject(a any) error
}
