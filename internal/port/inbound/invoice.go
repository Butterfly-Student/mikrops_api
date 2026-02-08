package inbound_port

type InvoiceHttpPort interface {
	List(a any) error
	Create(a any) error
	Get(a any) error
	Update(a any) error
	GenerateBulk(a any) error
}
