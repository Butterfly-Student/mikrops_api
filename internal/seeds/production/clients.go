package production

import (
	"fmt"

	"go-template/internal/model"
	"go-template/internal/seeds/runner"

	"gorm.io/gorm"
)

type ClientSeeder struct{}

func (s *ClientSeeder) Name() string {
	return "Production Clients"
}

func (s *ClientSeeder) Seed(db *gorm.DB) error {
	systemClient := &model.Client{
		ClientInput: model.ClientInput{
			Name: "system-client",
		},
	}

	var existingClient model.Client
	result := db.Where("name = ?", systemClient.Name).First(&existingClient)

	if result.Error == gorm.ErrRecordNotFound {
		if err := db.Create(systemClient).Error; err != nil {
			return fmt.Errorf("failed to create system client: %w", err)
		}
		fmt.Printf("[CLIENT SEED] Created system client: %s\n", systemClient.Name)
	} else if result.Error != nil {
		return fmt.Errorf("failed to check existing client: %w", result.Error)
	} else {
		fmt.Printf("[CLIENT SEED] System client already exists: %s\n", systemClient.Name)
	}

	return nil
}

func init() {
	runner.RegisterSeeder(&ClientSeeder{})
}
