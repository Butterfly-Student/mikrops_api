package ping

import (
	"context"
	"log"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
)

type PingDomain interface {
	StartPing(ctx context.Context, routerID string, req model.PingRequest) error
	StopPing(routerID string, address string) error
	SubscribeToPingResults(ctx context.Context, address string) (<-chan model.WebSocketMessage, error)
}

type domain struct {
	dbPort       outbound_port.DatabasePort
	cachePort    outbound_port.CachePort
	mikrotikPort outbound_port.MikrotikPort
}

func NewPingDomain(
	dbPort outbound_port.DatabasePort,
	cachePort outbound_port.CachePort,
	mikrotikPort outbound_port.MikrotikPort,
) PingDomain {
	return &domain{
		dbPort:       dbPort,
		cachePort:    cachePort,
		mikrotikPort: mikrotikPort,
	}
}

func (d *domain) getRouter(routerID string) (*model.MikrotikRouter, error) {
	return d.dbPort.Mikrotik().FindByID(routerID)
}

// Ping Operations

func (d *domain) StartPing(ctx context.Context, routerID string, req model.PingRequest) error {
	router, err := d.getRouter(routerID)
	if err != nil {
		return err
	}

	resultCh, err := d.mikrotikPort.Ping(ctx, router, req)
	if err != nil {
		return err
	}

	go func() {
		for result := range resultCh {
			msg := model.WebSocketMessage{
				Event: "ping_result",
				Data:  result,
			}
			if err := d.cachePort.PppoePubSub().PublishPingResult(req.Address, msg); err != nil {
				log.Printf("Failed to publish ping result for %s: %v", req.Address, err)
			}
		}
	}()

	return nil
}

func (d *domain) StopPing(routerID string, address string) error {
	return nil
}

func (d *domain) SubscribeToPingResults(ctx context.Context, address string) (<-chan model.WebSocketMessage, error) {
	return d.cachePort.PppoePubSub().SubscribeToPingResults(ctx, address)
}
