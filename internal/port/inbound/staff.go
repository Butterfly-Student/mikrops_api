package inbound_port

type StaffHttpPort interface {
	List(a any) error
	Create(a any) error
	Get(a any) error
	Update(a any) error
	Delete(a any) error
}
