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

func (s *CasbinSeeder) Seed(db *gorm.DB) error {
	fmt.Printf("[CASBIN TEST SEED] Starting seeder...\n")

	policies := []model.CasbinRule{
		{Ptype: "p", V0: "admin", V1: "/api/v1/*", V2: "*"},
		{Ptype: "p", V0: "admin", V1: "/api/v1/clients/*", V2: "*"},
		{Ptype: "p", V0: "admin", V1: "/api/v1/users/*", V2: "*"},
		{Ptype: "p", V0: "admin", V1: "/api/v1/mikrotik", V2: "*"},
		{Ptype: "p", V0: "admin", V1: "/api/v1/mikrotik/*", V2: "*"},
		{Ptype: "p", V0: "admin", V1: "/api/v1/pppoe/*", V2: "*"},
		{Ptype: "p", V0: "admin", V1: "/api/v1/queues/*", V2: "*"},

		{Ptype: "p", V0: "user", V1: "/api/v1/mikrotik", V2: "GET"},
		{Ptype: "p", V0: "user", V1: "/api/v1/mikrotik/*", V2: "GET"},
		{Ptype: "p", V0: "user", V1: "/api/v1/pppoe/secrets", V2: "GET"},
		{Ptype: "p", V0: "user", V1: "/api/v1/pppoe/secrets/*", V2: "GET"},
		{Ptype: "p", V0: "user", V1: "/api/v1/queues", V2: "GET"},
		{Ptype: "p", V0: "user", V1: "/api/v1/queues/*", V2: "GET"},

		{Ptype: "p", V0: "test-role", V1: "/api/v1/*", V2: "*"},

		{Ptype: "g", V0: "test-admin@example.com", V1: "admin"},
		{Ptype: "g", V0: "test-user@example.com", V1: "user"},
		{Ptype: "g", V0: "inactive-user@example.com", V1: "user"},
		{Ptype: "g", V0: "regular@example.com", V1: "user"},
	}

	for _, policy := range policies {
		var existingRule model.CasbinRule
		query := db.Where("ptype = ? AND v0 = ?", policy.Ptype, policy.V0)

		if policy.V1 != "" {
			query = query.Where("v1 = ?", policy.V1)
		}
		if policy.V2 != "" {
			query = query.Where("v2 = ?", policy.V2)
		}

		result := query.First(&existingRule)

		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(&policy).Error; err != nil {
				return fmt.Errorf("failed to create casbin rule: %w", err)
			}
			fmt.Printf("[CASBIN TEST SEED] Created rule: %s, %s, %s, %s\n",
				policy.Ptype, policy.V0, policy.V1, policy.V2)
		} else if result.Error != nil {
			return fmt.Errorf("failed to check existing casbin rule: %w", result.Error)
		}
	}

	fmt.Printf("[CASBIN TEST SEED] Completed successfully\n")
	return nil
}

func init() {
	runner.RegisterSeeder(&CasbinSeeder{})
}
