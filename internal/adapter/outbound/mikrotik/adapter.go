package mikrotik_outbound_adapter

import (
	"mikrops/internal/model"
	outbound_port "mikrops/internal/port/outbound"
)

type adapter struct{}

func NewAdapter() outbound_port.MikrotikPort {
	return &adapter{}
}

// Connection & Status
func (a *adapter) TestConnection(host string, port int, username, password string, useSSL bool) error {
	client := NewRouterOSClient(host, port, username, password)
	return client.TestConnection()
}

func (a *adapter) GetIdentity(host string, port int, username, password string, useSSL bool) (string, error) {
	client := NewRouterOSClient(host, port, username, password)
	return client.GetIdentity()
}

func (a *adapter) GetActiveConnections(host string, port int, username, password string, useSSL bool) ([]model.MikrotikConnection, error) {
	client := NewRouterOSClient(host, port, username, password)
	return client.GetActiveConnections()
}

// Interface Traffic
func (a *adapter) GetInterfaceTraffic(host string, port int, username, password string, useSSL bool, interfaceName string) (model.MikrotikTraffic, error) {
	client := NewRouterOSClient(host, port, username, password)
	return client.GetInterfaceTraffic(interfaceName)
}

// PPPoE Secrets (REST API)
func (a *adapter) ListPPPoESecrets(host string, restPort int, username, password string, useSSL bool) ([]model.MikrotikPPPoESecret, error) {
	client := NewRestClient(host, restPort, username, password, useSSL)
	return client.ListPPPoESecrets()
}

func (a *adapter) CreatePPPoESecret(host string, restPort int, username, password string, useSSL bool, data model.MikrotikPPPoESecretInput) error {
	client := NewRestClient(host, restPort, username, password, useSSL)
	return client.CreatePPPoESecret(data)
}

func (a *adapter) UpdatePPPoESecret(host string, restPort int, username, password string, useSSL bool, id string, data model.MikrotikPPPoESecretInput) error {
	client := NewRestClient(host, restPort, username, password, useSSL)
	return client.UpdatePPPoESecret(id, data)
}

func (a *adapter) DeletePPPoESecret(host string, restPort int, username, password string, useSSL bool, id string) error {
	client := NewRestClient(host, restPort, username, password, useSSL)
	return client.DeletePPPoESecret(id)
}

func (a *adapter) DisablePPPoESecret(host string, restPort int, username, password string, useSSL bool, id string) error {
	client := NewRestClient(host, restPort, username, password, useSSL)
	return client.DisablePPPoESecret(id)
}

func (a *adapter) EnablePPPoESecret(host string, restPort int, username, password string, useSSL bool, id string) error {
	client := NewRestClient(host, restPort, username, password, useSSL)
	return client.EnablePPPoESecret(id)
}

// PPPoE Profiles (REST API)
func (a *adapter) ListPPPoEProfiles(host string, restPort int, username, password string, useSSL bool) ([]model.MikrotikProfile, error) {
	client := NewRestClient(host, restPort, username, password, useSSL)
	return client.ListPPPoEProfiles()
}

func (a *adapter) CreatePPPoEProfile(host string, restPort int, username, password string, useSSL bool, data model.MikrotikProfileInput) error {
	client := NewRestClient(host, restPort, username, password, useSSL)
	return client.CreatePPPoEProfile(data)
}

// Hotspot Users (REST API)
func (a *adapter) ListHotspotUsers(host string, restPort int, username, password string, useSSL bool) ([]model.MikrotikHotspotUser, error) {
	client := NewRestClient(host, restPort, username, password, useSSL)
	return client.ListHotspotUsers()
}

func (a *adapter) CreateHotspotUser(host string, restPort int, username, password string, useSSL bool, data model.MikrotikHotspotUserInput) error {
	client := NewRestClient(host, restPort, username, password, useSSL)
	return client.CreateHotspotUser(data)
}

func (a *adapter) UpdateHotspotUser(host string, restPort int, username, password string, useSSL bool, id string, data model.MikrotikHotspotUserInput) error {
	client := NewRestClient(host, restPort, username, password, useSSL)
	return client.UpdateHotspotUser(id, data)
}

func (a *adapter) DeleteHotspotUser(host string, restPort int, username, password string, useSSL bool, id string) error {
	client := NewRestClient(host, restPort, username, password, useSSL)
	return client.DeleteHotspotUser(id)
}

// Simple Queue (REST API)
func (a *adapter) ListSimpleQueues(host string, restPort int, username, password string, useSSL bool) ([]model.MikrotikSimpleQueue, error) {
	client := NewRestClient(host, restPort, username, password, useSSL)
	return client.ListSimpleQueues()
}

func (a *adapter) CreateSimpleQueue(host string, restPort int, username, password string, useSSL bool, data model.MikrotikSimpleQueueInput) error {
	client := NewRestClient(host, restPort, username, password, useSSL)
	return client.CreateSimpleQueue(data)
}

func (a *adapter) UpdateSimpleQueue(host string, restPort int, username, password string, useSSL bool, id string, data model.MikrotikSimpleQueueInput) error {
	client := NewRestClient(host, restPort, username, password, useSSL)
	return client.UpdateSimpleQueue(id, data)
}

func (a *adapter) DeleteSimpleQueue(host string, restPort int, username, password string, useSSL bool, id string) error {
	client := NewRestClient(host, restPort, username, password, useSSL)
	return client.DeleteSimpleQueue(id)
}

// Bandwidth Monitoring (RouterOS API)
func (a *adapter) MonitorBandwidth(host string, port int, username, password string, useSSL bool, target string) (model.MikrotikBandwidth, error) {
	client := NewRouterOSClient(host, port, username, password)
	return client.MonitorBandwidth(target)
}
