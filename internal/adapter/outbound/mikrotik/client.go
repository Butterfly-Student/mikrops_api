package mikrotik_outbound_adapter

import (
	"crypto/tls"
	"fmt"
	"sync"
	"time"

	"github.com/go-routeros/routeros/v3"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
)

type mikrotikClientAdapter struct {
	clients map[uint]*routeros.Client
	mu      sync.Mutex
}

func NewMikrotikClientAdapter() outbound_port.MikrotikPort {
	return &mikrotikClientAdapter{
		clients: make(map[uint]*routeros.Client),
	}
}

func (a *mikrotikClientAdapter) getClient(router *model.MikrotikRouter) (*routeros.Client, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Check if client exists and is connected
	if client, ok := a.clients[router.ID]; ok {
		// Basic check if connection is alive (dummy command)
		// If dead, we reconnect
		_, err := client.Run("/system/identity/print")
		if err == nil {
			return client, nil
		}
		client.Close()
		delete(a.clients, router.ID)
	}

	// Connect
	// Assuming address is "ip:port" or just "ip" (default 8728)
	// For API-SSL (8729), we might need logic to detect or config
	client, err := routeros.Dial(router.Address, router.Username, router.Password)
	if err != nil {
		// Try TLS/SSL if plain fails? Or explicit config?
		// For now assuming plain API
		return nil, fmt.Errorf("failed to dial router: %w", err)
	}

	a.clients[router.ID] = client
	return client, nil
}
