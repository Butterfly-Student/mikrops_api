package inbound_port

type MikrotikHttpPort interface {
	// PPPoE Secrets
	ListPPPoESecrets(a any) error
	CreatePPPoESecret(a any) error
	UpdatePPPoESecret(a any) error
	DeletePPPoESecret(a any) error
	DisablePPPoESecret(a any) error
	EnablePPPoESecret(a any) error

	// PPPoE Profiles
	ListPPPoEProfiles(a any) error
	CreatePPPoEProfile(a any) error

	// Hotspot Users
	ListHotspotUsers(a any) error
	CreateHotspotUser(a any) error
	UpdateHotspotUser(a any) error
	DeleteHotspotUser(a any) error

	// Simple Queues
	ListSimpleQueues(a any) error
	CreateSimpleQueue(a any) error
	UpdateSimpleQueue(a any) error
	DeleteSimpleQueue(a any) error

	// Monitoring
	GetActiveConnections(a any) error
	GetInterfaceTraffic(a any) error
}
