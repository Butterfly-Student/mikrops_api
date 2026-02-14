package testing

import (
	"fmt"

	"go-template/internal/model"
	"go-template/internal/seeds/runner"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserSeeder struct{}

func (s *UserSeeder) Name() string {
	return "Testing Users"
}

func (s *UserSeeder) Seed(db *gorm.DB) error {
	testUsers := []*model.User{
		{
			Name:   "Test Admin",
			Email:  "test-admin@example.com",
			Role:   "admin",
			Status: "active",
		},
		{
			Name:   "Test User",
			Email:  "test-user@example.com",
			Role:   "user",
			Status: "active",
		},
		{
			Name:   "Inactive User",
			Email:  "inactive-user@example.com",
			Role:   "user",
			Status: "inactive",
		},
		{
			Name:   "Regular User",
			Email:  "regular@example.com",
			Role:   "user",
			Status: "active",
		},
	}

	for _, user := range testUsers {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("Test@123"), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password for %s: %w", user.Email, err)
		}
		user.Password = string(hashedPassword)

		var existingUser model.User
		result := db.Where("email = ?", user.Email).First(&existingUser)

		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(user).Error; err != nil {
				return fmt.Errorf("failed to create test user %s: %w", user.Email, err)
			}
			fmt.Printf("[USER TEST SEED] Created: %s\n", user.Email)
		} else if result.Error != nil {
			return fmt.Errorf("failed to check existing user %s: %w", user.Email, result.Error)
		} else {
			fmt.Printf("[USER TEST SEED] Already exists: %s\n", user.Email)
		}
	}

	return nil
}

func init() {
	runner.RegisterSeeder(&UserSeeder{})
}
