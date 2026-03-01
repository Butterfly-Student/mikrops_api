package fixtures

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"go-template/internal/model"
	"go-template/utils/hash"
)

type UserTestData struct{}

func NewUserTestData() *UserTestData {
	return &UserTestData{}
}

// ValidUserInput returns a valid user input for testing
func (u *UserTestData) ValidUserInput() model.AdminUserInput {
	return model.AdminUserInput{
		FullName: "Test User",
		Email:    "test@example.com",
		Password: "TestPassword@123",
		Role:     "cs",
		IsActive: boolPtr(true),
	}
}

// ValidUser returns a valid user model for testing
func (u *UserTestData) ValidUser() model.AdminUser {
	now := time.Now()
	hashedPassword, _ := hash.HashPassword("TestPassword@123")
	isActive := true

	return model.AdminUser{
		ID:           uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
		FullName:     "Test User",
		Email:        "test@example.com",
		PasswordHash: hashedPassword,
		Role:         model.AdminRoleCS,
		IsActive:     &isActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// AdminUser returns an admin user for testing
func (u *UserTestData) AdminUser() model.AdminUser {
	now := time.Now()
	hashedPassword, _ := hash.HashPassword("AdminPassword@123")
	isActive := true

	return model.AdminUser{
		ID:           uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
		FullName:     "Admin User",
		Email:        "admin@example.com",
		PasswordHash: hashedPassword,
		Role:         model.AdminRoleAdmin,
		IsActive:     &isActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// InactiveUser returns an inactive user for testing
func (u *UserTestData) InactiveUser() model.AdminUser {
	now := time.Now()
	hashedPassword, _ := hash.HashPassword("InactivePassword@123")
	isInactive := false

	return model.AdminUser{
		ID:           uuid.MustParse("550e8400-e29b-41d4-a716-446655440003"),
		FullName:     "Inactive User",
		Email:        "inactive@example.com",
		PasswordHash: hashedPassword,
		Role:         model.AdminRoleCS,
		IsActive:     &isInactive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// ValidUserFilter returns a valid user filter for testing
func (u *UserTestData) ValidUserFilter() model.AdminUserFilter {
	return model.AdminUserFilter{
		Emails: []string{"test@example.com"},
	}
}

// MultipleUsers returns multiple users for testing
func (u *UserTestData) MultipleUsers(count int) []model.AdminUser {
	users := make([]model.AdminUser, count)
	now := time.Now()
	isActive := true

	for i := 0; i < count; i++ {
		hashedPassword, _ := hash.HashPassword(fmt.Sprintf("Password%d@123", i+1))
		role := model.AdminRoleCS
		if i == 0 {
			role = model.AdminRoleAdmin
		}

		users[i] = model.AdminUser{
			ID:           uuid.MustParse(fmt.Sprintf("550e8400-e29b-41d4-a716-446655440%03d", i+1)),
			FullName:     fmt.Sprintf("User %d", i+1),
			Email:        fmt.Sprintf("user%d@example.com", i+1),
			PasswordHash: hashedPassword,
			Role:         role,
			IsActive:     &isActive,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
	}

	return users
}

// MultipleUserInputs returns multiple user inputs for testing
func (u *UserTestData) MultipleUserInputs(count int) []model.AdminUserInput {
	inputs := make([]model.AdminUserInput, count)
	isActive := true

	for i := 0; i < count; i++ {
		role := "cs"
		if i == 0 {
			role = "admin"
		}

		inputs[i] = model.AdminUserInput{
			FullName: fmt.Sprintf("User %d", i+1),
			Email:    fmt.Sprintf("user%d@example.com", i+1),
			Password: fmt.Sprintf("Password%d@123", i+1),
			Role:     role,
			IsActive: &isActive,
		}
	}

	return inputs
}

// UserWithRole returns a user with a specific role
func (u *UserTestData) UserWithRole(role string) model.AdminUser {
	now := time.Now()
	hashedPassword, _ := hash.HashPassword("TestPassword@123")
	isActive := true

	return model.AdminUser{
		ID:           uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
		FullName:     fmt.Sprintf("%s User", role),
		Email:        fmt.Sprintf("%s@example.com", role),
		PasswordHash: hashedPassword,
		Role:         model.AdminUserRole(role),
		IsActive:     &isActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// UserWithStatus returns a user with a specific status
func (u *UserTestData) UserWithStatus(isActive bool) model.AdminUser {
	now := time.Now()
	hashedPassword, _ := hash.HashPassword("TestPassword@123")

	return model.AdminUser{
		ID:           uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
		FullName:     "Test User",
		Email:        "test@example.com",
		PasswordHash: hashedPassword,
		Role:         model.AdminRoleCS,
		IsActive:     &isActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}


