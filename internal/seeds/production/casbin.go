package production

import (
	"fmt"

	"go-template/internal/model"
	"go-template/internal/seeds/runner"

	"gorm.io/gorm"
)

type CasbinSeeder struct{}

func (s *CasbinSeeder) Name() string {
	return "Production Casbin Rules"
}

// Seed inserts Casbin policy (p) rules for each role and grouping (g) rules
// that map each existing user's numeric ID to their role.
//
// Roles:
//   - admin : full access to all protected routes (any HTTP method)
//   - user  : operational staff — can manage customers, invoices, payments;
//             read-only access to PPPoE, queues, interfaces, IP pools,
//             bandwidth profiles, and MikroTik router configs
//
// The matcher uses keyMatch so patterns like "/pppoe/secrets/*" match
// "/pppoe/secrets/123". Method value "*" means any HTTP method.
func (s *CasbinSeeder) Seed(db *gorm.DB) error {
	fmt.Printf("[CASBIN SEED] Starting production seeder...\n")

	policies := []model.CasbinRule{
		// ─────────────────────────────────────────────────────────────────────
		// ADMIN — full access to all protected routes
		// ─────────────────────────────────────────────────────────────────────

		// Own profile & auth actions
		{Ptype: "p", V0: "admin", V1: "/user/profile", V2: "*"},
		{Ptype: "p", V0: "admin", V1: "/user/change-password", V2: "*"},
		{Ptype: "p", V0: "admin", V1: "/user/logout", V2: "*"},

		// Ping management
		{Ptype: "p", V0: "admin", V1: "/ping", V2: "*"},
		{Ptype: "p", V0: "admin", V1: "/ping/*", V2: "*"},

		// PPPoE (global)
		{Ptype: "p", V0: "admin", V1: "/pppoe/*", V2: "*"},

		// Queues (global)
		{Ptype: "p", V0: "admin", V1: "/queues", V2: "*"},
		{Ptype: "p", V0: "admin", V1: "/queues/*", V2: "*"},

		// Interface monitoring (global)
		{Ptype: "p", V0: "admin", V1: "/interfaces/monitor", V2: "*"},
		{Ptype: "p", V0: "admin", V1: "/interfaces/monitor/*", V2: "*"},

		// IP Pools (global)
		{Ptype: "p", V0: "admin", V1: "/ip-pools", V2: "*"},
		{Ptype: "p", V0: "admin", V1: "/ip-pools/*", V2: "*"},

		// Bandwidth Profiles (global)
		{Ptype: "p", V0: "admin", V1: "/bandwidth-profiles", V2: "*"},
		{Ptype: "p", V0: "admin", V1: "/bandwidth-profiles/*", V2: "*"},

		// Customers (global reads)
		{Ptype: "p", V0: "admin", V1: "/customers", V2: "*"},
		{Ptype: "p", V0: "admin", V1: "/customers/*", V2: "*"},

		// Invoices
		{Ptype: "p", V0: "admin", V1: "/invoices", V2: "*"},
		{Ptype: "p", V0: "admin", V1: "/invoices/*", V2: "*"},

		// Payments
		{Ptype: "p", V0: "admin", V1: "/payments", V2: "*"},
		{Ptype: "p", V0: "admin", V1: "/payments/*", V2: "*"},

		// MikroTik: router CRUD + all /mikrotik/:router_id/* operations
		{Ptype: "p", V0: "admin", V1: "/mikrotik", V2: "*"},
		{Ptype: "p", V0: "admin", V1: "/mikrotik/*", V2: "*"},

		// Registration management
		{Ptype: "p", V0: "admin", V1: "/registrations", V2: "*"},
		{Ptype: "p", V0: "admin", V1: "/registrations/*", V2: "*"},

		// ─────────────────────────────────────────────────────────────────────
		// USER (operational staff) — day-to-day ISP operations
		// ─────────────────────────────────────────────────────────────────────

		// Ping: operational staff can start/stop pings for diagnostics
		{Ptype: "p", V0: "user", V1: "/ping", V2: "POST"},
		{Ptype: "p", V0: "user", V1: "/ping/*", V2: "DELETE"},

		// Own profile & auth actions
		{Ptype: "p", V0: "user", V1: "/user/profile", V2: "GET"},
		{Ptype: "p", V0: "user", V1: "/user/profile", V2: "PUT"},
		{Ptype: "p", V0: "user", V1: "/user/change-password", V2: "POST"},
		{Ptype: "p", V0: "user", V1: "/user/logout", V2: "POST"},

		// PPPoE: read-only (monitoring only, cannot manage secrets/profiles)
		{Ptype: "p", V0: "user", V1: "/pppoe/secrets", V2: "GET"},
		{Ptype: "p", V0: "user", V1: "/pppoe/secrets/*", V2: "GET"},
		{Ptype: "p", V0: "user", V1: "/pppoe/profiles", V2: "GET"},
		{Ptype: "p", V0: "user", V1: "/pppoe/profiles/*", V2: "GET"},
		{Ptype: "p", V0: "user", V1: "/pppoe/sessions/*", V2: "GET"},

		// Queues: read-only
		{Ptype: "p", V0: "user", V1: "/queues", V2: "GET"},
		{Ptype: "p", V0: "user", V1: "/queues/*", V2: "GET"},

		// Interface monitoring: can start/stop monitoring streams
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

		// Customers (global): read-only
		// Writes go through /mikrotik/:router_id/customers/* (covered below)
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

		// Registration management: user can list/view and approve/reject
		{Ptype: "p", V0: "user", V1: "/registrations", V2: "GET"},
		{Ptype: "p", V0: "user", V1: "/registrations/*", V2: "GET"},
		{Ptype: "p", V0: "user", V1: "/registrations/*", V2: "POST"},

		// MikroTik routers: read-only (cannot create/update/delete router configs)
		{Ptype: "p", V0: "user", V1: "/mikrotik", V2: "GET"},

		// MikroTik router operations: full access for operational tasks
		// Covers /mikrotik/:router_id/customers/*, /pppoe/*, /hotspot/*, etc.
		{Ptype: "p", V0: "user", V1: "/mikrotik/*", V2: "GET"},
		{Ptype: "p", V0: "user", V1: "/mikrotik/*", V2: "POST"},
		{Ptype: "p", V0: "user", V1: "/mikrotik/*", V2: "PUT"},
		{Ptype: "p", V0: "user", V1: "/mikrotik/*", V2: "DELETE"},
	}

	for _, rule := range policies {
		if err := upsertPolicyRule(db, rule); err != nil {
			return err
		}
	}

	// ─── Grouping rules: map each user's numeric ID → their role ──────────────
	// The RBAC middleware uses fmt.Sprintf("%d", userID) as the Casbin subject,
	// so groupings MUST use the numeric ID (not email).
	var users []model.User
	if err := db.Find(&users).Error; err != nil {
		return fmt.Errorf("[CASBIN SEED] failed to query users: %w", err)
	}
	for _, u := range users {
		if u.Role == "" {
			continue
		}
		gRule := model.CasbinRule{
			Ptype: "g",
			V0:    u.ID.String(),
			V1:    string(u.Role),
		}
		if err := upsertGroupRule(db, gRule); err != nil {
			return err
		}
		fmt.Printf("[CASBIN SEED] Assigned role %q to user %s (id=%s)\n", u.Role, u.Email, u.ID.String())
	}

	fmt.Printf("[CASBIN SEED] Completed successfully\n")
	return nil
}

// upsertPolicyRule inserts a policy (p) rule only if it does not already exist.
func upsertPolicyRule(db *gorm.DB, rule model.CasbinRule) error {
	var existing model.CasbinRule
	result := db.Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?",
		rule.Ptype, rule.V0, rule.V1, rule.V2).First(&existing)
	if result.Error == gorm.ErrRecordNotFound {
		if err := db.Create(&rule).Error; err != nil {
			return fmt.Errorf("[CASBIN SEED] failed to create policy rule (%s %s %s): %w",
				rule.V0, rule.V1, rule.V2, err)
		}
	}
	return nil
}

// upsertGroupRule inserts a grouping (g) rule only if it does not already exist.
func upsertGroupRule(db *gorm.DB, rule model.CasbinRule) error {
	var existing model.CasbinRule
	result := db.Where("ptype = ? AND v0 = ? AND v1 = ?",
		rule.Ptype, rule.V0, rule.V1).First(&existing)
	if result.Error == gorm.ErrRecordNotFound {
		if err := db.Create(&rule).Error; err != nil {
			return fmt.Errorf("[CASBIN SEED] failed to create group rule (%s → %s): %w",
				rule.V0, rule.V1, err)
		}
	}
	return nil
}

func init() {
	runner.RegisterSeeder(&CasbinSeeder{})
}
