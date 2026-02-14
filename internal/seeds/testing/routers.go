package testing

import (
	"fmt"

	"go-template/internal/model"
	"go-template/internal/seeds/runner"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MikrotikSeeder struct{}

func (s *MikrotikSeeder) Name() string {
	return "Testing MikroTik Routers"
}

func (s *MikrotikSeeder) Seed(db *gorm.DB) error {
	apiPort := 8728
	restPort := 80
	useSSL := false
	isActive := true

	testRouters := []*model.MikrotikRouter{
		{
			ID:       uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
			Name:     "Test Router 1",
			Address:  "192.168.1.1:8728",
			ApiPort:  &apiPort,
			RestPort: &restPort,
			Username: "admin",
			Password: "test-password",
			UseSSL:   &useSSL,
			IsActive: &isActive,
		},
		{
			ID:       uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
			Name:     "Test Router 2",
			Address:  "192.168.1.2:8728",
			ApiPort:  &apiPort,
			RestPort: &restPort,
			Username: "testadmin",
			Password: "test-password-2",
			UseSSL:   &useSSL,
			IsActive: &isActive,
		},
		{
			ID:       uuid.MustParse("550e8400-e29b-41d4-a716-446655440003"),
			Name:     "Inactive Test Router",
			Address:  "192.168.1.3:8728",
			ApiPort:  &apiPort,
			RestPort: &restPort,
			Username: "inactive-admin",
			Password: "inactive-password",
			UseSSL:   &useSSL,
			IsActive: func(b bool) *bool { return &b }(false),
		},
	}

	for _, router := range testRouters {
		var existingRouter model.MikrotikRouter
		result := db.Where("id = ?", router.ID).First(&existingRouter)

		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(router).Error; err != nil {
				return fmt.Errorf("failed to create test router %s: %w", router.Name, err)
			}
			fmt.Printf("[MIKROTIK TEST SEED] Created: %s\n", router.Name)
		} else if result.Error != nil {
			return fmt.Errorf("failed to check existing router %s: %w", router.Name, result.Error)
		} else {
			fmt.Printf("[MIKROTIK TEST SEED] Already exists: %s\n", router.Name)
		}
	}

	return nil
}

func init() {
	runner.RegisterSeeder(&MikrotikSeeder{})
}
