package inbound_port

type MikrotikSyncLogHttpPort interface {
	List(a any) error
	Get(a any) error
}
