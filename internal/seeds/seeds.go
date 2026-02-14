package seeds

import (
	"fmt"
	"log"
	"os"

	_ "go-template/internal/seeds/production"
	"go-template/internal/seeds/runner"
	_ "go-template/internal/seeds/testing"

	"gorm.io/gorm"
)

func Run(db *gorm.DB) error {
	config := runner.GetSeedConfig()
	log.Printf("[SEED] Environment: %s", config.Env)
	log.Printf("[SEED] Entities: %v", config.Entities)

	allSeeders := runner.GetRegisteredSeeders()
	var filteredSeeders []runner.Seeder

	for _, seeder := range allSeeders {
		seederName := seeder.Name()

		shouldRun := false

		// Filter by environment
		switch config.Env {
		case runner.Production:
			if !contains(seederName, "Testing") {
				shouldRun = true
			}
		case runner.Testing:
			if !contains(seederName, "Production") {
				shouldRun = true
			}
		default:
			// Development or staging - use both or only production seeders
			shouldRun = true
		}

		if !shouldRun {
			continue
		}

		// Filter by entity type
		if !runner.ShouldSeedEntity("all", config.Entities) {
			entityMatched := false
			for _, entity := range config.Entities {
				if entity == "users" && (seederName == "Production Users" || seederName == "Testing Users") {
					entityMatched = true
					break
				}
				if entity == "clients" && (seederName == "Production Clients" || seederName == "Testing Clients") {
					entityMatched = true
					break
				}
				if entity == "routers" && (seederName == "Production MikroTik Routers" || seederName == "Testing MikroTik Routers") {
					entityMatched = true
					break
				}
				if entity == "casbin" && (seederName == "Production Casbin Rules" || seederName == "Testing Casbin Rules") {
					entityMatched = true
					break
				}
			}
			shouldRun = entityMatched
		}

		if shouldRun {
			filteredSeeders = append(filteredSeeders, seeder)
		}
	}

	if len(filteredSeeders) == 0 {
		log.Println("[SEED] No seeders matched the criteria")
		return nil
	}

	log.Printf("[SEED] Running %d seeders", len(filteredSeeders))
	results, err := runner.RunInTransaction(db, filteredSeeders)
	if err != nil {
		log.Printf("[SEED] Failed: %v", err)
		for _, result := range results {
			if !result.Success {
				log.Printf("[SEED] Failed seeder: %s - Error: %v", result.Name, result.Error)
			}
		}
		return fmt.Errorf("seeding failed: %w", err)
	}

	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		}
	}

	log.Printf("[SEED] Completed: %d/%d successful", successCount, len(results))
	return nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			indexOf(s, substr) >= 0)))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func RunForTesting(db *gorm.DB) error {
	oldEnv := os.Getenv("SEED_ENV")
	defer os.Setenv("SEED_ENV", oldEnv)

	os.Setenv("SEED_ENV", "testing")

	return Run(db)
}
