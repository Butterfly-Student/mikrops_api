package production

import (
	"fmt"
	"os"

	"go-template/internal/model"
	"go-template/internal/seeds/runner"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MikrotikSeeder struct{}

func (s *MikrotikSeeder) Name() string {
	return "Production MikroTik Routers"
}

func (s *MikrotikSeeder) Seed(db *gorm.DB) error {
	seedRouters := os.Getenv("SEED_ROUTERS")
	if seedRouters == "" || seedRouters == "false" {
		fmt.Println("[MIKROTIK SEED] Skipping router seeding (not configured)")
		return nil
	}

	routerName := os.Getenv("SEED_ROUTER_NAME")
	if routerName == "" {
		routerName = "Default Router"
	}

	routerAddress := os.Getenv("SEED_ROUTER_ADDRESS")
	if routerAddress == "" {
		fmt.Println("[MIKROTIK SEED] No router address provided, skipping")
		return nil
	}

	routerUsername := os.Getenv("SEED_ROUTER_USERNAME")
	if routerUsername == "" {
		fmt.Println("[MIKROTIK SEED] No router username provided, skipping")
		return nil
	}

	routerPassword := os.Getenv("SEED_ROUTER_PASSWORD")
	if routerPassword == "" {
		fmt.Println("[MIKROTIK SEED] No router password provided, skipping")
		return nil
	}

	router := &model.MikrotikRouter{
		ID:       uuid.New(),
		Name:     routerName,
		Address:  routerAddress,
		Username: routerUsername,
		Password: routerPassword,
		IsActive: func(b bool) *bool { return &b }(true),
	}

	var existingRouter model.MikrotikRouter
	result := db.Where("name = ?", router.Name).First(&existingRouter)

	if result.Error == gorm.ErrRecordNotFound {
		if err := db.Create(router).Error; err != nil {
			return fmt.Errorf("failed to create mikrotik router: %w", err)
		}
		fmt.Printf("[MIKROTIK SEED] Created router: %s\n", router.Name)
	} else if result.Error != nil {
		return fmt.Errorf("failed to check existing router: %w", result.Error)
	} else {
		fmt.Printf("[MIKROTIK SEED] Router already exists: %s\n", router.Name)
	}

	return nil
}

func init() {
	runner.RegisterSeeder(&MikrotikSeeder{})
}
