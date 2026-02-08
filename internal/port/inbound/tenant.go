package inbound_port

type TenantHttpPort interface {
	Get(a any) error
	Update(a any) error
}
