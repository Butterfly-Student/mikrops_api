package inbound_port

type CustomerRegistrationHttpPort interface {
	List(a any) error
	Get(a any) error
	Create(a any) error
	Approve(a any) error
	Reject(a any) error
	PublicSubmit(a any) error // Public registration endpoint without authentication
}
