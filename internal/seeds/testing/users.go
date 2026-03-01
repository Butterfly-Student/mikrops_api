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
	isActive := true
	isInactive := false

	testUsers := []*model.AdminUser{
		{
			FullName: "Test Admin",
			Email:    "test-admin@example.com",
			Role:     model.AdminRoleAdmin,
			IsActive: &isActive,
		},
		{
			FullName: "Test User",
			Email:    "test-user@example.com",
			Role:     model.AdminRoleCS,
			IsActive: &isActive,
		},
		{
			FullName: "Inactive User",
			Email:    "inactive-user@example.com",
			Role:     model.AdminRoleCS,
			IsActive: &isInactive,
		},
		{
			FullName: "Regular User",
			Email:    "regular@example.com",
			Role:     model.AdminRoleCS,
			IsActive: &isActive,
		},
	}

	for _, user := range testUsers {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("Test@123"), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password for %s: %w", user.Email, err)
		}
		user.PasswordHash = string(hashedPassword)

		var existingUser model.AdminUser
		result := db.Where("email = ?", user.Email).First(&existingUser)

		var targetID string
		var targetRole string

		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(user).Error; err != nil {
				return fmt.Errorf("failed to create test user %s: %w", user.Email, err)
			}
			fmt.Printf("[USER TEST SEED] Created: %s (id=%s)\n", user.Email, user.ID.String())
			targetID = user.ID.String()
			targetRole = string(user.Role)
		} else if result.Error != nil {
			return fmt.Errorf("failed to check existing user %s: %w", user.Email, result.Error)
		} else {
			fmt.Printf("[USER TEST SEED] Already exists: %s\n", user.Email)
			targetID = existingUser.ID.String()
			targetRole = string(existingUser.Role)
		}

		// Add Casbin grouping immediately so RBAC works without running the
		// Casbin seeder separately. The RBAC middleware uses the user ID
		// as the Casbin subject.
		if targetRole != "" {
			gRule := model.CasbinRule{
				Ptype: "g",
				V0:    targetID,
				V1:    targetRole,
			}
			if err := upsertTestGroupRule(db, gRule); err != nil {
				return fmt.Errorf("failed to assign casbin role for %s: %w", user.Email, err)
			}
		}
	}

	return nil
}

func init() {
	runner.RegisterSeeder(&UserSeeder{})
}
