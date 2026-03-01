package production

import (
	"fmt"
	"os"

	"go-template/internal/model"
	"go-template/internal/seeds/runner"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ClientSeeder struct{}

func (s *ClientSeeder) Name() string {
	return "Production Clients"
}

func (s *ClientSeeder) Seed(db *gorm.DB) error {
	apiKey := os.Getenv("SYSTEM_API_KEY")
	if apiKey == "" {
		apiKey = uuid.New().String()
	}

	isActive := true
	systemClient := &model.AdminUser{
		ID:           uuid.New(),
		FullName:     "system-client",
		Email:        "system@mikrotik.internal",
		PasswordHash: "$2a$10$disabled", // account login disabled, API key only
		Role:         model.AdminRoleReadOnly,
		BearerKey:    &apiKey,
		IsActive:     &isActive,
	}

	var existingClient model.AdminUser
	result := db.Where("email = ?", systemClient.Email).First(&existingClient)

	if result.Error == gorm.ErrRecordNotFound {
		if err := db.Create(systemClient).Error; err != nil {
			return fmt.Errorf("failed to create system client: %w", err)
		}
		fmt.Printf("[CLIENT SEED] Created system client: %s\n", systemClient.FullName)
	} else if result.Error != nil {
		return fmt.Errorf("failed to check existing client: %w", result.Error)
	} else {
		fmt.Printf("[CLIENT SEED] System client already exists: %s\n", existingClient.FullName)
	}

	return nil
}

func init() {
	runner.RegisterSeeder(&ClientSeeder{})
}
