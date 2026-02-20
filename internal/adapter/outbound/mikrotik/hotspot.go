package mikrotik_outbound_adapter

import (
	"fmt"
	"sync"

	routeros "github.com/go-routeros/routeros/v3"
	"github.com/google/uuid"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
	"go-template/pkg/hotspot"
)

// hotspotMikrotikAdapter implements HotspotPort.
// It maintains its own RouterOS connection pool keyed by router UUID.
type hotspotMikrotikAdapter struct {
	clients map[uuid.UUID]*routeros.Client
	mu      sync.Mutex
}

// NewHotspotAdapter creates a new HotspotPort adapter.
func NewHotspotAdapter() outbound_port.HotspotPort {
	return &hotspotMikrotikAdapter{
		clients: make(map[uuid.UUID]*routeros.Client),
	}
}

func (a *hotspotMikrotikAdapter) getClient(router *model.MikrotikRouter) (*routeros.Client, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if client, ok := a.clients[router.ID]; ok {
		_, err := client.Run("/system/identity/print")
		if err == nil {
			return client, nil
		}
		client.Close()
		delete(a.clients, router.ID)
	}

	client, err := routeros.Dial(router.Address, router.Username, router.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to dial router: %w", err)
	}

	a.clients[router.ID] = client
	return client, nil
}

// GetHotspotClient returns a hotspot.Client connected to the given router.
func (a *hotspotMikrotikAdapter) GetHotspotClient(router *model.MikrotikRouter) (hotspot.Client, error) {
	rosClient, err := a.getClient(router)
	if err != nil {
		return nil, err
	}
	return hotspot.NewClient(0, rosClient), nil
}
