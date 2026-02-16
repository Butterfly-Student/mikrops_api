package mikrotik

import (
	"fmt"
	"time"

	"go-template/internal/model"
)

type mikrotikDomain struct {
	dbPort  MikrotikDatabasePort
	apiPort MikrotikApiPort
}

type MikrotikDatabasePort interface {
	Create(router *model.MikrotikRouter) error
	FindByID(id string) (*model.MikrotikRouter, error)
	FindAll() ([]model.MikrotikRouter, error)
	Update(router *model.MikrotikRouter) error
	Delete(id string) error
	SetActive(id string) error
	GetActive() (*model.MikrotikRouter, error)
	DeactivateAll() error
	UpdateLastSeen(id string) error
	TestConnection(router *model.MikrotikRouter) error
}

type MikrotikApiPort interface {
	TestConnection(router *model.MikrotikRouter) error
}

func NewMikrotikDomain(dbPort MikrotikDatabasePort, apiPort MikrotikApiPort) MikrotikDomain {
	return &mikrotikDomain{
		dbPort:  dbPort,
		apiPort: apiPort,
	}
}

func (d *mikrotikDomain) Create(name, address, username, password string) (*model.MikrotikRouter, error) {
	if name == "" || address == "" || username == "" || password == "" {
		return nil, ErrRouterInvalid
	}

	router := &model.MikrotikRouter{
		Name:     name,
		Address:  address,
		Username: username,
		Password: password,
		IsActive: boolPtr(false),
	}

	if err := d.dbPort.Create(router); err != nil {
		return nil, fmt.Errorf("failed to create router: %w", err)
	}

	return router, nil
}

func (d *mikrotikDomain) FindByID(id string) (*model.MikrotikRouter, error) {
	router, err := d.dbPort.FindByID(id)
	if err != nil {
		return nil, ErrRouterNotFound
	}
	return router, nil
}

func (d *mikrotikDomain) FindAll() ([]model.MikrotikRouter, error) {
	return d.dbPort.FindAll()
}

func (d *mikrotikDomain) Update(id string, updates map[string]interface{}) (*model.MikrotikRouter, error) {
	router, err := d.FindByID(id)
	if err != nil {
		return nil, err
	}

	if v, ok := updates["name"]; ok && v != nil {
		if name, ok := v.(string); ok && name != "" {
			router.Name = name
		}
	}
	if v, ok := updates["address"]; ok && v != nil {
		if address, ok := v.(string); ok && address != "" {
			router.Address = address
		}
	}
	if v, ok := updates["username"]; ok && v != nil {
		if username, ok := v.(string); ok && username != "" {
			router.Username = username
		}
	}
	if v, ok := updates["password"]; ok && v != nil {
		if password, ok := v.(string); ok && password != "" {
			router.Password = password
		}
	}
	if v, ok := updates["api_port"]; ok && v != nil {
		if port, ok := v.(int); ok {
			router.ApiPort = &port
		}
	}
	if v, ok := updates["rest_port"]; ok && v != nil {
		if port, ok := v.(int); ok {
			router.RestPort = &port
		}
	}
	if v, ok := updates["use_ssl"]; ok && v != nil {
		if ssl, ok := v.(bool); ok {
			router.UseSSL = &ssl
		}
	}
	if v, ok := updates["password_encrypted"]; ok && v != nil {
		if encrypted, ok := v.(string); ok {
			router.PasswordEncrypted = &encrypted
		}
	}

	if err := d.dbPort.Update(router); err != nil {
		return nil, fmt.Errorf("failed to update router: %w", err)
	}

	return router, nil
}

func (d *mikrotikDomain) Delete(id string) error {
	if _, err := d.FindByID(id); err != nil {
		return err
	}

	if err := d.dbPort.Delete(id); err != nil {
		return fmt.Errorf("failed to delete router: %w", err)
	}

	return nil
}

func (d *mikrotikDomain) SetActive(id string) (*model.MikrotikRouter, error) {
	router, err := d.FindByID(id)
	if err != nil {
		return nil, err
	}

	if router.IsRouterActive() {
		return router, nil
	}

	if err := d.dbPort.SetActive(id); err != nil {
		return nil, fmt.Errorf("failed to set active router: %w", err)
	}

	router.IsActive = boolPtr(true)
	return router, nil
}

func (d *mikrotikDomain) GetActive() (*model.MikrotikRouter, error) {
	router, err := d.dbPort.GetActive()
	if err != nil {
		return nil, ErrNoActiveRouter
	}
	return router, nil
}

func (d *mikrotikDomain) DeactivateAll() error {
	return d.dbPort.DeactivateAll()
}

func (d *mikrotikDomain) UpdateLastSeen(id string) error {
	if _, err := d.FindByID(id); err != nil {
		return err
	}

	return d.dbPort.UpdateLastSeen(id)
}

func (d *mikrotikDomain) TestConnection(id string) error {
	router, err := d.FindByID(id)
	if err != nil {
		return err
	}

	if err := d.apiPort.TestConnection(router); err != nil {
		return ErrConnectionFailed
	}

	now := time.Now()
	router.LastSeenAt = &now
	if err := d.dbPort.Update(router); err != nil {
		return fmt.Errorf("failed to update last seen: %w", err)
	}

	return nil
}

func boolPtr(b bool) *bool {
	return &b
}
