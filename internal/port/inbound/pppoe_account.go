package inbound_port

type PppoeAccountHttpPort interface {
	List(a any) error
	Get(a any) error
	Create(a any) error
	Update(a any) error
	Delete(a any) error
	Isolate(a any) error
	Restore(a any) error
}
