package inbound_port

type ActivityLogHttpPort interface {
	List(a any) error
	Get(a any) error
}
