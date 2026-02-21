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

	adminUser := &model.User{
		Name:     "Admin",
		Email:    email,
		Password: string(hashedPassword),
		Role:     "admin",
		Status:   "active",
	}

	var existingUser model.User
	result := db.Where("email = ?", adminUser.Email).First(&existingUser)

	if result.Error == gorm.ErrRecordNotFound {
		if err := db.Create(adminUser).Error; err != nil {
			return fmt.Errorf("failed to create admin user: %w", err)
		}
		fmt.Printf("[USER SEED] Created admin user: %s (id=%d)\n", email, adminUser.ID)

		// Add Casbin grouping immediately so RBAC works without running the
		// Casbin seeder separately. The RBAC middleware uses the numeric user ID
		// as the Casbin subject, so we must use that — not the email.
		gRule := model.CasbinRule{
			Ptype: "g",
			V0:    fmt.Sprintf("%d", adminUser.ID),
			V1:    adminUser.Role,
		}
		if err := upsertGroupRule(db, gRule); err != nil {
			return fmt.Errorf("failed to assign casbin role for admin user: %w", err)
		}
		fmt.Printf("[USER SEED] Assigned Casbin role %q to user id=%d\n", adminUser.Role, adminUser.ID)
	} else if result.Error != nil {
		return fmt.Errorf("failed to check existing user: %w", result.Error)
	} else {
		fmt.Printf("[USER SEED] Admin user already exists: %s\n", email)

		// Ensure grouping exists even for pre-existing users
		gRule := model.CasbinRule{
			Ptype: "g",
			V0:    fmt.Sprintf("%d", existingUser.ID),
			V1:    existingUser.Role,
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
