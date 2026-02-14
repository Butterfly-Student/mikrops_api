package ippool

import (
	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
)

type IpPoolDomain interface {
	CreateIpPool(routerID string, pool model.IpPool) error
	GetIpPool(routerID string, id string) (*model.IpPool, error)
	ListIpPools(routerID string) ([]model.IpPool, error)
	UpdateIpPool(routerID string, pool model.IpPool) error
	DeleteIpPool(routerID string, id string) error
}

type domain struct {
	dbPort       outbound_port.DatabasePort
	mikrotikPort outbound_port.MikrotikPort
}

func NewIpPoolDomain(
	dbPort outbound_port.DatabasePort,
	mikrotikPort outbound_port.MikrotikPort,
) IpPoolDomain {
	return &domain{
		dbPort:       dbPort,
		mikrotikPort: mikrotikPort,
	}
}

func (d *domain) getRouter(routerID string) (*model.MikrotikRouter, error) {
	return d.dbPort.Mikrotik().FindByID(routerID)
}

// CRUD Operations

func (d *domain) CreateIpPool(routerID string, pool model.IpPool) error {
	router, err := d.getRouter(routerID)
	if err != nil {
		return err
	}
	return d.mikrotikPort.CreateIpPool(router, &pool)
}

func (d *domain) GetIpPool(routerID string, id string) (*model.IpPool, error) {
	router, err := d.getRouter(routerID)
	if err != nil {
		return nil, err
	}
	return d.mikrotikPort.GetIpPool(router, id)
}

func (d *domain) ListIpPools(routerID string) ([]model.IpPool, error) {
	router, err := d.getRouter(routerID)
	if err != nil {
		return nil, err
	}
	return d.mikrotikPort.ListIpPools(router)
}

func (d *domain) UpdateIpPool(routerID string, pool model.IpPool) error {
	router, err := d.getRouter(routerID)
	if err != nil {
		return err
	}
	return d.mikrotikPort.UpdateIpPool(router, &pool)
}

func (d *domain) DeleteIpPool(routerID string, id string) error {
	router, err := d.getRouter(routerID)
	if err != nil {
		return err
	}
	return d.mikrotikPort.DeleteIpPool(router, id)
}
