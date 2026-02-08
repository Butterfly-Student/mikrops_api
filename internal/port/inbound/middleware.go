package inbound_port

type MiddlewareHttpPort interface {
	InternalAuth(a any) error
	ClientAuth(a any) error
	StaffAuth(a any) error
	CustomerAuth(a any) error
	RequirePermission(resource, action string) func(a any) error
}
