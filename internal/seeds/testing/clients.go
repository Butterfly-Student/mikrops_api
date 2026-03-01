package testing

import (
	"fmt"

	"go-template/internal/model"
	"go-template/internal/seeds/runner"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type ClientSeeder struct{}

func (s *ClientSeeder) Name() string {
	return "Testing Clients"
}

func (s *ClientSeeder) Seed(db *gorm.DB) error {
	fmt.Printf("[CLIENT TEST SEED] Starting seeder...\n")

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("Test@123"), bcrypt.DefaultCost)
	isActive := true

	testClients := []*model.AdminUser{
		{
			ID:           uuid.New(),
			FullName:     "test-client-1",
			Email:        "test-client-1@example.com",
			PasswordHash: string(hashedPassword),
			Role:         model.AdminRoleAdmin,
			IsActive:     &isActive,
		},
		{
			ID:           uuid.New(),
			FullName:     "test-client-2",
			Email:        "test-client-2@example.com",
			PasswordHash: string(hashedPassword),
			Role:         model.AdminRoleCS,
			IsActive:     &isActive,
		},
		{
			ID:           uuid.New(),
			FullName:     "integration-test-client",
			Email:        "integration-test-client@example.com",
			PasswordHash: string(hashedPassword),
			Role:         model.AdminRoleAdmin,
			IsActive:     &isActive,
		},
	}

	for _, client := range testClients {
		fmt.Printf("[CLIENT TEST SEED] Processing client: %s\n", client.FullName)

		var existingClient model.AdminUser
		result := db.Where("email = ?", client.Email).First(&existingClient)

		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(client).Error; err != nil {
				return fmt.Errorf("failed to create test client %s: %w", client.FullName, err)
			}
			fmt.Printf("[CLIENT TEST SEED] Created: %s\n", client.FullName)
		} else if result.Error != nil {
			return fmt.Errorf("failed to check existing client %s: %w", client.FullName, result.Error)
		} else {
			fmt.Printf("[CLIENT TEST SEED] Already exists: %s\n", client.FullName)
		}
	}

	return nil
}

func init() {
	runner.RegisterSeeder(&ClientSeeder{})
}
