package postgres_outbound_adapter

import (
	"fmt"
	"time"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"

	"github.com/go-routeros/routeros/v3"
	"gorm.io/gorm"
)

type mikrotikAdapter struct {
	db *gorm.DB
}

func NewMikrotikAdapter(db *gorm.DB) outbound_port.MikrotikDatabasePort {
	return &mikrotikAdapter{db: db}
}

func (a *mikrotikAdapter) Create(router *model.MikrotikRouter) error {
	return a.db.Create(router).Error
}

func (a *mikrotikAdapter) FindByID(id string) (*model.MikrotikRouter, error) {
	var router model.MikrotikRouter
	err := a.db.Where("id = ?", id).First(&router).Error
	if err != nil {
		return nil, err
	}
	return &router, nil
}

func (a *mikrotikAdapter) FindAll() ([]model.MikrotikRouter, error) {
	var routers []model.MikrotikRouter
	err := a.db.Find(&routers).Error
	return routers, err
}

func (a *mikrotikAdapter) Update(router *model.MikrotikRouter) error {
	return a.db.Save(router).Error
}

func (a *mikrotikAdapter) Delete(id string) error {
	return a.db.Where("id = ?", id).Delete(&model.MikrotikRouter{}).Error
}

// Firewall methods - These are currently no-op as firewall rules are not stored in the database
// They are managed directly on the MikroTik router via the MikrotikPort interface
func (a *mikrotikAdapter) AddFirewallRule(router *model.MikrotikRouter, rule model.FirewallRule) error {
	// Firewall rules are managed on the router itself, not in the database
	// This method is here to satisfy the interface but does nothing
	return nil
}

func (a *mikrotikAdapter) RemoveFirewallRule(router *model.MikrotikRouter, rule model.FirewallRule) error {
	// Firewall rules are managed on the router itself, not in the database
	// This method is here to satisfy the interface but does nothing
	return nil
}

func (a *mikrotikAdapter) ListFirewallRules(router *model.MikrotikRouter) ([]model.FirewallRule, error) {
	// Firewall rules are managed on the router itself, not in the database
	// This method is here to satisfy of interface but returns empty slice
	return []model.FirewallRule{}, nil
}

// SetActive sets a router as active and deactivates all other routers
// This ensures only one router can be active at a time (single-active enforcement)
func (a *mikrotikAdapter) SetActive(id string) error {
	tx := a.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Deactivate all routers
	if err := tx.Exec("UPDATE mikrotik_routers SET is_active = false WHERE is_active = true").Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to deactivate all routers: %w", err)
	}

	// Activate the specified router
	if err := tx.Model(&model.MikrotikRouter{}).Where("id = ?", id).Update("is_active", true).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to activate router %s: %w", id, err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetActive retrieves the currently active router
// Returns error if no active router exists
func (a *mikrotikAdapter) GetActive() (*model.MikrotikRouter, error) {
	var router model.MikrotikRouter
	err := a.db.Where("is_active = ?", true).First(&router).Error
	if err != nil {
		return nil, err
	}
	return &router, nil
}

// DeactivateAll deactivates all routers
func (a *mikrotikAdapter) DeactivateAll() error {
	return a.db.Exec("UPDATE mikrotik_routers SET is_active = false").Error
}

// UpdateLastSeen updates the last_seen_at timestamp for a router
func (a *mikrotikAdapter) UpdateLastSeen(id string) error {
	now := time.Now()
	return a.db.Model(&model.MikrotikRouter{}).Where("id = ?", id).Update("last_seen_at", now).Error
}

// TestConnection tests the connection to a MikroTik router
func (a *mikrotikAdapter) TestConnection(router *model.MikrotikRouter) error {
	client, err := routeros.Dial(router.Address, router.Username, router.Password)
	if err != nil {
		return fmt.Errorf("failed to dial router: %w", err)
	}
	defer client.Close()

	// Test command
	_, err = client.Run("/system/identity/print")
	if err != nil {
		return fmt.Errorf("failed to run test command: %w", err)
	}

	return nil
}
