package outbound_port

import "mikrops/internal/model"

//go:generate mockgen -source=mikrotik.go -destination=./../../../tests/mocks/port/mock_mikrotik.go
type MikrotikPort interface {
	// Connection & Status
	TestConnection(host string, port int, username, password string, useSSL bool) error
	GetIdentity(host string, port int, username, password string, useSSL bool) (string, error)
	GetActiveConnections(host string, port int, username, password string, useSSL bool) ([]model.MikrotikConnection, error)

	// Interface Traffic
	GetInterfaceTraffic(host string, port int, username, password string, useSSL bool, interfaceName string) (model.MikrotikTraffic, error)

	// PPPoE Secrets (REST API)
	ListPPPoESecrets(host string, restPort int, username, password string, useSSL bool) ([]model.MikrotikPPPoESecret, error)
	CreatePPPoESecret(host string, restPort int, username, password string, useSSL bool, data model.MikrotikPPPoESecretInput) error
	UpdatePPPoESecret(host string, restPort int, username, password string, useSSL bool, id string, data model.MikrotikPPPoESecretInput) error
	DeletePPPoESecret(host string, restPort int, username, password string, useSSL bool, id string) error
	DisablePPPoESecret(host string, restPort int, username, password string, useSSL bool, id string) error
	EnablePPPoESecret(host string, restPort int, username, password string, useSSL bool, id string) error

	// PPPoE Profiles (REST API)
	ListPPPoEProfiles(host string, restPort int, username, password string, useSSL bool) ([]model.MikrotikProfile, error)
	CreatePPPoEProfile(host string, restPort int, username, password string, useSSL bool, data model.MikrotikProfileInput) error

	// Hotspot Users (REST API)
	ListHotspotUsers(host string, restPort int, username, password string, useSSL bool) ([]model.MikrotikHotspotUser, error)
	CreateHotspotUser(host string, restPort int, username, password string, useSSL bool, data model.MikrotikHotspotUserInput) error
	UpdateHotspotUser(host string, restPort int, username, password string, useSSL bool, id string, data model.MikrotikHotspotUserInput) error
	DeleteHotspotUser(host string, restPort int, username, password string, useSSL bool, id string) error

	// Simple Queue (REST API)
	ListSimpleQueues(host string, restPort int, username, password string, useSSL bool) ([]model.MikrotikSimpleQueue, error)
	CreateSimpleQueue(host string, restPort int, username, password string, useSSL bool, data model.MikrotikSimpleQueueInput) error
	UpdateSimpleQueue(host string, restPort int, username, password string, useSSL bool, id string, data model.MikrotikSimpleQueueInput) error
	DeleteSimpleQueue(host string, restPort int, username, password string, useSSL bool, id string) error

	// Bandwidth Monitoring (RouterOS API)
	MonitorBandwidth(host string, port int, username, password string, useSSL bool, target string) (model.MikrotikBandwidth, error)
}
