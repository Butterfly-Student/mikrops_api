package mikrotik_router

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/palantir/stacktrace"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
)

// MikrotikRouterDomain defines the interface for MikroTik router management
type MikrotikRouterDomain interface {
	Create(ctx context.Context, input model.MikrotikRouterInput) (*model.MikrotikRouter, error)
	GetByID(ctx context.Context, id string) (*model.MikrotikRouter, error)
	List(ctx context.Context) ([]model.MikrotikRouter, error)
	Update(ctx context.Context, id string, input model.MikrotikRouterInput) (*model.MikrotikRouter, error)
	Delete(ctx context.Context, id string) error
	TestConnection(ctx context.Context, id string) (*model.MikrotikTestResult, error)
}

type domain struct {
	dbPort       outbound_port.DatabasePort
	mikrotikPort outbound_port.MikrotikPort
}

func NewMikrotikRouterDomain(
	dbPort outbound_port.DatabasePort,
	mikrotikPort outbound_port.MikrotikPort,
) MikrotikRouterDomain {
	return &domain{
		dbPort:       dbPort,
		mikrotikPort: mikrotikPort,
	}
}

func (d *domain) Create(ctx context.Context, input model.MikrotikRouterInput) (*model.MikrotikRouter, error) {
	router := &model.MikrotikRouter{
		ID:       uuid.New(),
		Name:     input.Name,
		Address:  input.Address,
		ApiPort:  input.ApiPort,
		Username: input.Username,
		Password: input.Password,
		UseSSL:   input.UseSSL,
		IsActive: input.IsActive,
	}

	if err := d.dbPort.Mikrotik().Create(router); err != nil {
		return nil, stacktrace.Propagate(err, "failed to create mikrotik router")
	}

	return router, nil
}

func (d *domain) GetByID(ctx context.Context, id string) (*model.MikrotikRouter, error) {
	router, err := d.dbPort.Mikrotik().FindByID(id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get mikrotik router by id")
	}
	return router, nil
}

func (d *domain) List(ctx context.Context) ([]model.MikrotikRouter, error) {
	routers, err := d.dbPort.Mikrotik().FindAll()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to list mikrotik routers")
	}
	return routers, nil
}

func (d *domain) Update(ctx context.Context, id string, input model.MikrotikRouterInput) (*model.MikrotikRouter, error) {
	router, err := d.dbPort.Mikrotik().FindByID(id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to find mikrotik router")
	}

	router.Name = input.Name
	router.Address = input.Address
	router.Username = input.Username
	router.Password = input.Password
	if input.ApiPort != nil {
		router.ApiPort = input.ApiPort
	}
	if input.UseSSL != nil {
		router.UseSSL = input.UseSSL
	}
	if input.IsActive != nil {
		router.IsActive = input.IsActive
	}

	if err := d.dbPort.Mikrotik().Update(router); err != nil {
		return nil, stacktrace.Propagate(err, "failed to update mikrotik router")
	}

	return router, nil
}

func (d *domain) Delete(ctx context.Context, id string) error {
	if _, err := d.dbPort.Mikrotik().FindByID(id); err != nil {
		return stacktrace.Propagate(err, "mikrotik router not found")
	}
	if err := d.dbPort.Mikrotik().Delete(id); err != nil {
		return stacktrace.Propagate(err, "failed to delete mikrotik router")
	}
	return nil
}

func (d *domain) TestConnection(ctx context.Context, id string) (*model.MikrotikTestResult, error) {
	router, err := d.dbPort.Mikrotik().FindByID(id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "mikrotik router not found")
	}

	// Try to list secrets as a connectivity test (lightweight operation)
	_, err = d.mikrotikPort.ListSecrets(router)
	if err != nil {
		return &model.MikrotikTestResult{
			Success: false,
			Message: fmt.Sprintf("Connection failed: %s", err.Error()),
		}, nil
	}

	result := &model.MikrotikTestResult{
		Success: true,
		Message: "Connection successful",
	}

	// Fill in optional fields from router
	if router.RouterOSVersion != nil {
		result.RouterOSVersion = *router.RouterOSVersion
	}
	if router.Identity != nil {
		result.Identity = *router.Identity
	}

	return result, nil
}
