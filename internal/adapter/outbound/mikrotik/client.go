package mikrotik_outbound_adapter

import (
	"fmt"
	"sync"

	"github.com/go-routeros/routeros/v3"
	"github.com/google/uuid"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
)

type mikrotikClientAdapter struct {
	clients       map[uuid.UUID]*routeros.Client
	streamManager *StreamManager
	mu            sync.Mutex
}

func NewMikrotikClientAdapter() outbound_port.MikrotikPort {
	return &mikrotikClientAdapter{
		clients:       make(map[uuid.UUID]*routeros.Client),
		streamManager: NewStreamManager(),
	}
}

func (a *mikrotikClientAdapter) getClient(router *model.MikrotikRouter) (*routeros.Client, error) {
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

func (a *mikrotikClientAdapter) TestConnection(router *model.MikrotikRouter) error {
	client, err := a.getClient(router)
	if err != nil {
		return err
	}

	_, err = client.Run("/system/identity/print")
	return err
}
