package inbound_port

type SubscriptionHttpPort interface {
	List(a any) error
	Create(a any) error
	Get(a any) error
	Update(a any) error
	Suspend(a any) error
	Activate(a any) error
	Cancel(a any) error
	SetVacation(a any) error
}
