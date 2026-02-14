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

func (s *CasbinSeeder) Seed(db *gorm.DB) error {
	fmt.Printf("[CASBIN SEED] Starting production seeder...\n")

	policies := []model.CasbinRule{
		{Ptype: "p", V0: "admin", V1: "/api/v1/*", V2: "*"},
		{Ptype: "p", V0: "admin", V1: "/api/v1/clients/*", V2: "*"},
		{Ptype: "p", V0: "admin", V1: "/api/v1/users/*", V2: "*"},
		{Ptype: "p", V0: "admin", V1: "/api/v1/mikrotik/*", V2: "*"},
		{Ptype: "p", V0: "admin", V1: "/api/v1/pppoe/*", V2: "*"},
		{Ptype: "p", V0: "admin", V1: "/api/v1/queues/*", V2: "*"},

		{Ptype: "p", V0: "user", V1: "/api/v1/mikrotik/*", V2: "GET"},
		{Ptype: "p", V0: "user", V1: "/api/v1/pppoe/secrets", V2: "GET"},
		{Ptype: "p", V0: "user", V1: "/api/v1/queues", V2: "GET"},

		{Ptype: "g", V0: "admin@mikrotik.local", V1: "admin"},
	}

	for _, policy := range policies {
		var existingRule model.CasbinRule
		result := db.Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?",
			policy.Ptype, policy.V0, policy.V1, policy.V2).First(&existingRule)

		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(&policy).Error; err != nil {
				return fmt.Errorf("failed to create casbin rule: %w", err)
			}
			fmt.Printf("[CASBIN SEED] Created rule: %s, %s, %s, %s\n",
				policy.Ptype, policy.V0, policy.V1, policy.V2)
		}
	}

	return nil
}

func init() {
	runner.RegisterSeeder(&CasbinSeeder{})
}
