package testing

import (
	"fmt"

	"go-template/internal/model"
	"go-template/internal/seeds/runner"

	"gorm.io/gorm"
)

type ClientSeeder struct{}

func (s *ClientSeeder) Name() string {
	return "Testing Clients"
}

func (s *ClientSeeder) Seed(db *gorm.DB) error {
	fmt.Printf("[CLIENT TEST SEED] Starting seeder...\n")

	testClients := []*model.Client{
		{
			ClientInput: model.ClientInput{
				Name: "test-client-1",
			},
		},
		{
			ClientInput: model.ClientInput{
				Name: "test-client-2",
			},
		},
		{
			ClientInput: model.ClientInput{
				Name: "integration-test-client",
			},
		},
	}

	for _, client := range testClients {
		fmt.Printf("[CLIENT TEST SEED] Processing client: %s\n", client.Name)

		var existingClient model.Client
		result := db.Where("name = ?", client.Name).First(&existingClient)

		if result.Error == gorm.ErrRecordNotFound {
			model.ClientPrepare(&client.ClientInput)
			if err := db.Create(client).Error; err != nil {
				return fmt.Errorf("failed to create test client %s: %w", client.Name, err)
			}
			fmt.Printf("[CLIENT TEST SEED] Created: %s with bearer key\n", client.Name)
		} else if result.Error != nil {
			return fmt.Errorf("failed to check existing client %s: %w", client.Name, result.Error)
		} else {
			fmt.Printf("[CLIENT TEST SEED] Already exists: %s\n", client.Name)
		}
	}

	return nil
}

func init() {
	runner.RegisterSeeder(&ClientSeeder{})
}
