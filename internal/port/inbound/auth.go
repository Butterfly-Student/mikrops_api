package inbound_port

type AuthHttpPort interface {
	StaffLogin(a any) error
	StaffRefresh(a any) error
	CustomerLogin(a any) error
	CustomerRefresh(a any) error
}
