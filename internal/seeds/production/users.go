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
		fmt.Printf("[USER SEED] Created admin user: %s\n", email)
	} else if result.Error != nil {
		return fmt.Errorf("failed to check existing user: %w", result.Error)
	} else {
		fmt.Printf("[USER SEED] Admin user already exists: %s\n", email)
	}

	return nil
}

func init() {
	runner.RegisterSeeder(&UserSeeder{})
}
