package production

import (
	"fmt"
	"os"

	"go-template/internal/model"
	"go-template/internal/seeds/runner"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserSeeder struct{}

func (s *UserSeeder) Name() string {
	return "Production Users"
}

func (s *UserSeeder) Seed(db *gorm.DB) error {
	email := os.Getenv("ADMIN_EMAIL")
	if email == "" {
		email = "admin@mikrotik.local"
	}

	password := os.Getenv("ADMIN_PASSWORD")
	if password == "" {
		password = "Admin@123"
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	isActive := true
	adminUser := &model.AdminUser{
		FullName:     "Administrator",
		Email:        email,
		PasswordHash: string(hashedPassword),
		Role:         model.AdminRoleSuperAdmin,
		IsActive:     &isActive,
	}

	var existingUser model.AdminUser
	result := db.Where("email = ?", adminUser.Email).First(&existingUser)

	if result.Error == gorm.ErrRecordNotFound {
		if err := db.Create(adminUser).Error; err != nil {
			return fmt.Errorf("failed to create admin user: %w", err)
		}
		fmt.Printf("[USER SEED] Created admin user: %s (id=%s)\n", email, adminUser.ID)

		gRule := model.CasbinRule{
			Ptype: "g",
			V0:    adminUser.ID.String(),
			V1:    string(adminUser.Role),
		}
		if err := upsertGroupRule(db, gRule); err != nil {
			return fmt.Errorf("failed to assign casbin role for admin user: %w", err)
		}
		fmt.Printf("[USER SEED] Assigned Casbin role %q to user id=%s\n", adminUser.Role, adminUser.ID)
	} else if result.Error != nil {
		return fmt.Errorf("failed to check existing user: %w", result.Error)
	} else {
		fmt.Printf("[USER SEED] Admin user already exists: %s\n", email)

		gRule := model.CasbinRule{
			Ptype: "g",
			V0:    existingUser.ID.String(),
			V1:    string(existingUser.Role),
		}
		if err := upsertGroupRule(db, gRule); err != nil {
			return fmt.Errorf("failed to ensure casbin role for admin user: %w", err)
		}
	}

	return nil
}

func init() {
	runner.RegisterSeeder(&UserSeeder{})
}
