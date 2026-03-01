package testing

import (
	"fmt"

	"go-template/internal/model"
	"go-template/internal/seeds/runner"

	"gorm.io/gorm"
)

type CasbinSeeder struct{}

func (s *CasbinSeeder) Name() string {
	return "Testing Casbin Rules"
}

// Seed inserts the same policy rules as the production seeder, then looks up
// all seeded test users by email and creates grouping (g) rules using their
// numeric IDs — matching how the RBAC middleware resolves subjects.
func (s *CasbinSeeder) Seed(db *gorm.DB) error {
	fmt.Printf("[CASBIN TEST SEED] Starting seeder...\n")

	policies := []model.CasbinRule{
		// ─────────────────────────────────────────────────────────────────────
		// ADMIN — full access to all protected routes
		// ─────────────────────────────────────────────────────────────────────
		{Ptype: "p", V0: "admin", V1: "/user/profile", V2: "*"},
		{Ptype: "p", V0: "admin", V1: "/user/change-password", V2: "*"},
		{Ptype: "p", V0: "admin", V1: "/user/logout", V2: "*"},
		{Ptype: "p", V0: "admin", V1: "/ping", V2: "*"},
		{Ptype: "p", V0: "admin", V1: "/ping/*", V2: "*"},
		{Ptype: "p", V0: "admin", V1: "/pppoe/*", V2: "*"},
		{Ptype: "p", V0: "admin", V1: "/queues", V2: "*"},
		{Ptype: "p", V0: "admin", V1: "/queues/*", V2: "*"},
		{Ptype: "p", V0: "admin", V1: "/interfaces/monitor", V2: "*"},
		{Ptype: "p", V0: "admin", V1: "/interfaces/monitor/*", V2: "*"},
		{Ptype: "p", V0: "admin", V1: "/ip-pools", V2: "*"},
		{Ptype: "p", V0: "admin", V1: "/ip-pools/*", V2: "*"},
		{Ptype: "p", V0: "admin", V1: "/bandwidth-profiles", V2: "*"},
		{Ptype: "p", V0: "admin", V1: "/bandwidth-profiles/*", V2: "*"},
		{Ptype: "p", V0: "admin", V1: "/customers", V2: "*"},
		{Ptype: "p", V0: "admin", V1: "/customers/*", V2: "*"},
		{Ptype: "p", V0: "admin", V1: "/invoices", V2: "*"},
		{Ptype: "p", V0: "admin", V1: "/invoices/*", V2: "*"},
		{Ptype: "p", V0: "admin", V1: "/payments", V2: "*"},
		{Ptype: "p", V0: "admin", V1: "/payments/*", V2: "*"},
		{Ptype: "p", V0: "admin", V1: "/mikrotik", V2: "*"},
		{Ptype: "p", V0: "admin", V1: "/mikrotik/*", V2: "*"},

		// ─────────────────────────────────────────────────────────────────────
		// USER (operational staff)
		// ─────────────────────────────────────────────────────────────────────
		{Ptype: "p", V0: "user", V1: "/user/profile", V2: "GET"},
		{Ptype: "p", V0: "user", V1: "/user/profile", V2: "PUT"},
		{Ptype: "p", V0: "user", V1: "/user/change-password", V2: "POST"},
		{Ptype: "p", V0: "user", V1: "/user/logout", V2: "POST"},

		// PPPoE: read-only
		{Ptype: "p", V0: "user", V1: "/pppoe/secrets", V2: "GET"},
		{Ptype: "p", V0: "user", V1: "/pppoe/secrets/*", V2: "GET"},
		{Ptype: "p", V0: "user", V1: "/pppoe/profiles", V2: "GET"},
		{Ptype: "p", V0: "user", V1: "/pppoe/profiles/*", V2: "GET"},
		{Ptype: "p", V0: "user", V1: "/pppoe/sessions/*", V2: "GET"},

		// Queues: read-only
		{Ptype: "p", V0: "user", V1: "/queues", V2: "GET"},
		{Ptype: "p", V0: "user", V1: "/queues/*", V2: "GET"},

		// Interface monitoring: can start/stop
		{Ptype: "p", V0: "user", V1: "/interfaces/monitor", V2: "POST"},
		{Ptype: "p", V0: "user", V1: "/interfaces/monitor/*", V2: "POST"},
		{Ptype: "p", V0: "user", V1: "/interfaces/monitor", V2: "DELETE"},
		{Ptype: "p", V0: "user", V1: "/interfaces/monitor/*", V2: "DELETE"},

		// IP Pools: read-only
		{Ptype: "p", V0: "user", V1: "/ip-pools", V2: "GET"},
		{Ptype: "p", V0: "user", V1: "/ip-pools/*", V2: "GET"},

		// Bandwidth Profiles: read-only
		{Ptype: "p", V0: "user", V1: "/bandwidth-profiles", V2: "GET"},
		{Ptype: "p", V0: "user", V1: "/bandwidth-profiles/*", V2: "GET"},

		// Customers: read-only (writes via /mikrotik/:router_id)
		{Ptype: "p", V0: "user", V1: "/customers", V2: "GET"},
		{Ptype: "p", V0: "user", V1: "/customers/*", V2: "GET"},

		// Invoices: full management
		{Ptype: "p", V0: "user", V1: "/invoices", V2: "GET"},
		{Ptype: "p", V0: "user", V1: "/invoices", V2: "POST"},
		{Ptype: "p", V0: "user", V1: "/invoices/*", V2: "GET"},
		{Ptype: "p", V0: "user", V1: "/invoices/*", V2: "POST"},
		{Ptype: "p", V0: "user", V1: "/invoices/*", V2: "PUT"},
		{Ptype: "p", V0: "user", V1: "/invoices/*", V2: "DELETE"},

		// Payments: full management
		{Ptype: "p", V0: "user", V1: "/payments", V2: "GET"},
		{Ptype: "p", V0: "user", V1: "/payments", V2: "POST"},
		{Ptype: "p", V0: "user", V1: "/payments/*", V2: "GET"},
		{Ptype: "p", V0: "user", V1: "/payments/*", V2: "POST"},
		{Ptype: "p", V0: "user", V1: "/payments/*", V2: "PUT"},
		{Ptype: "p", V0: "user", V1: "/payments/*", V2: "DELETE"},

		// MikroTik routers: read-only list
		{Ptype: "p", V0: "user", V1: "/mikrotik", V2: "GET"},

		// MikroTik router operations: full operational access
		{Ptype: "p", V0: "user", V1: "/mikrotik/*", V2: "GET"},
		{Ptype: "p", V0: "user", V1: "/mikrotik/*", V2: "POST"},
		{Ptype: "p", V0: "user", V1: "/mikrotik/*", V2: "PUT"},
		{Ptype: "p", V0: "user", V1: "/mikrotik/*", V2: "DELETE"},
	}

	for _, rule := range policies {
		if err := upsertTestPolicyRule(db, rule); err != nil {
			return err
		}
	}

	// ─── Grouping rules: look up test users by email, assign by numeric ID ────
	// Matches how RBAC middleware resolves subjects: fmt.Sprintf("%d", userID)
	testUserEmails := []string{
		"test-admin@example.com",
		"test-user@example.com",
		"inactive-user@example.com",
		"regular@example.com",
	}

	for _, email := range testUserEmails {
		var u model.User
		if err := db.Where("email = ?", email).First(&u).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				fmt.Printf("[CASBIN TEST SEED] User %s not found, skipping grouping\n", email)
				continue
			}
			return fmt.Errorf("[CASBIN TEST SEED] failed to find user %s: %w", email, err)
		}
		if u.Role == "" {
			continue
		}
		gRule := model.CasbinRule{
			Ptype: "g",
			V0:    u.ID.String(),
			V1:    string(u.Role),
		}
		if err := upsertTestGroupRule(db, gRule); err != nil {
			return err
		}
		fmt.Printf("[CASBIN TEST SEED] Assigned role %q to user %s (id=%s)\n", u.Role, email, u.ID.String())
	}

	fmt.Printf("[CASBIN TEST SEED] Completed successfully\n")
	return nil
}

func upsertTestPolicyRule(db *gorm.DB, rule model.CasbinRule) error {
	var existing model.CasbinRule
	result := db.Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?",
		rule.Ptype, rule.V0, rule.V1, rule.V2).First(&existing)
	if result.Error == gorm.ErrRecordNotFound {
		if err := db.Create(&rule).Error; err != nil {
			return fmt.Errorf("[CASBIN TEST SEED] failed to create policy rule (%s %s %s): %w",
				rule.V0, rule.V1, rule.V2, err)
		}
	} else if result.Error != nil {
		return fmt.Errorf("[CASBIN TEST SEED] db error: %w", result.Error)
	}
	return nil
}

func upsertTestGroupRule(db *gorm.DB, rule model.CasbinRule) error {
	var existing model.CasbinRule
	result := db.Where("ptype = ? AND v0 = ? AND v1 = ?",
		rule.Ptype, rule.V0, rule.V1).First(&existing)
	if result.Error == gorm.ErrRecordNotFound {
		if err := db.Create(&rule).Error; err != nil {
			return fmt.Errorf("[CASBIN TEST SEED] failed to create group rule (%s → %s): %w",
				rule.V0, rule.V1, err)
		}
	} else if result.Error != nil {
		return fmt.Errorf("[CASBIN TEST SEED] db error: %w", result.Error)
	}
	return nil
}

func init() {
	runner.RegisterSeeder(&CasbinSeeder{})
}
