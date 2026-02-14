package iface

import (
	"context"
	"log"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
)

type InterfaceDomain interface {
	// Monitoring
	StartMonitoring(ctx context.Context, routerID string) error
	StartMonitoringByName(ctx context.Context, routerID string, interfaceName string) error
	StopMonitoring(routerID string) error
	StopMonitoringByName(routerID string, interfaceName string) error

	// Subscriptions
	SubscribeToStats(ctx context.Context) (<-chan model.WebSocketMessage, error)
	SubscribeToStatsByName(ctx context.Context, interfaceName string) (<-chan model.WebSocketMessage, error)
}

type domain struct {
	dbPort       outbound_port.DatabasePort
	cachePort    outbound_port.CachePort
	mikrotikPort outbound_port.MikrotikPort
}

func NewInterfaceDomain(
	dbPort outbound_port.DatabasePort,
	cachePort outbound_port.CachePort,
	mikrotikPort outbound_port.MikrotikPort,
) InterfaceDomain {
	return &domain{
		dbPort:       dbPort,
		cachePort:    cachePort,
		mikrotikPort: mikrotikPort,
	}
}

func (d *domain) getRouter(routerID string) (*model.MikrotikRouter, error) {
	return d.dbPort.Mikrotik().FindByID(routerID)
}

// Monitoring

func (d *domain) StartMonitoring(ctx context.Context, routerID string) error {
	router, err := d.getRouter(routerID)
	if err != nil {
		return err
	}

	statsCh, err := d.mikrotikPort.MonitorAllInterfaces(ctx, router)
	if err != nil {
		return err
	}

	go func() {
		for stats := range statsCh {
			msg := model.WebSocketMessage{
				Event: "interface_stats",
				Data:  stats,
			}
			if err := d.cachePort.PppoePubSub().PublishInterfaceStats(msg); err != nil {
				log.Printf("Failed to publish interface stats: %v", err)
			}
		}
	}()

	return nil
}

func (d *domain) StartMonitoringByName(ctx context.Context, routerID string, interfaceName string) error {
	router, err := d.getRouter(routerID)
	if err != nil {
		return err
	}

	statsCh, err := d.mikrotikPort.MonitorInterface(ctx, router, interfaceName)
	if err != nil {
		return err
	}

	go func() {
		for stats := range statsCh {
			msg := model.WebSocketMessage{
				Event: "interface_stats",
				Data:  stats,
			}
			if err := d.cachePort.PppoePubSub().PublishInterfaceStatsByName(interfaceName, msg); err != nil {
				log.Printf("Failed to publish interface stats for %s: %v", interfaceName, err)
			}
		}
	}()

	return nil
}

func (d *domain) StopMonitoring(routerID string) error {
	return nil
}

func (d *domain) StopMonitoringByName(routerID string, interfaceName string) error {
	return nil
}

// Subscriptions

func (d *domain) SubscribeToStats(ctx context.Context) (<-chan model.WebSocketMessage, error) {
	return d.cachePort.PppoePubSub().SubscribeToInterfaceStats(ctx)
}

func (d *domain) SubscribeToStatsByName(ctx context.Context, interfaceName string) (<-chan model.WebSocketMessage, error) {
	return d.cachePort.PppoePubSub().SubscribeToInterfaceStatsByName(ctx, interfaceName)
}
