package inbound_port

type TenantSettingHttpPort interface {
	Get(a any) error
	Upsert(a any) error
}
