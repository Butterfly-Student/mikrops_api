package fixtures

import (
	"fmt"
	"time"

	"go-template/internal/model"
	"go-template/utils/hash"
)

type UserTestData struct{}

func NewUserTestData() *UserTestData {
	return &UserTestData{}
}

// ValidUserInput returns a valid user input for testing
func (u *UserTestData) ValidUserInput() model.UserInput {
	return model.UserInput{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "TestPassword@123",
		Role:     "user",
		Status:   "active",
	}
}

// ValidUser returns a valid user model for testing
func (u *UserTestData) ValidUser() model.User {
	now := time.Now()
	hashedPassword, _ := hash.HashPassword("TestPassword@123")

	return model.User{
		ID:        1,
		Name:      "Test User",
		Email:     "test@example.com",
		Password:  hashedPassword,
		Role:      "user",
		Status:    "active",
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// AdminUser returns an admin user for testing
func (u *UserTestData) AdminUser() model.User {
	now := time.Now()
	hashedPassword, _ := hash.HashPassword("AdminPassword@123")

	return model.User{
		ID:        2,
		Name:      "Admin User",
		Email:     "admin@example.com",
		Password:  hashedPassword,
		Role:      "admin",
		Status:    "active",
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// InactiveUser returns an inactive user for testing
func (u *UserTestData) InactiveUser() model.User {
	now := time.Now()
	hashedPassword, _ := hash.HashPassword("InactivePassword@123")

	return model.User{
		ID:        3,
		Name:      "Inactive User",
		Email:     "inactive@example.com",
		Password:  hashedPassword,
		Role:      "user",
		Status:    "inactive",
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// ValidUserFilter returns a valid user filter for testing
func (u *UserTestData) ValidUserFilter() model.UserFilter {
	return model.UserFilter{
		Email: "test@example.com",
	}
}

// MultipleUsers returns multiple users for testing
func (u *UserTestData) MultipleUsers(count int) []model.User {
	users := make([]model.User, count)
	now := time.Now()

	for i := 0; i < count; i++ {
		hashedPassword, _ := hash.HashPassword(fmt.Sprintf("Password%d@123", i+1))
		role := "user"
		if i == 0 {
			role = "admin"
		}

		users[i] = model.User{
			ID:        uint(i + 1),
			Name:      fmt.Sprintf("User %d", i+1),
			Email:     fmt.Sprintf("user%d@example.com", i+1),
			Password:  hashedPassword,
			Role:      role,
			Status:    "active",
			CreatedAt: now,
			UpdatedAt: now,
		}
	}

	return users
}

// MultipleUserInputs returns multiple user inputs for testing
func (u *UserTestData) MultipleUserInputs(count int) []model.UserInput {
	inputs := make([]model.UserInput, count)

	for i := 0; i < count; i++ {
		role := "user"
		if i == 0 {
			role = "admin"
		}

		inputs[i] = model.UserInput{
			Name:     fmt.Sprintf("User %d", i+1),
			Email:    fmt.Sprintf("user%d@example.com", i+1),
			Password: fmt.Sprintf("Password%d@123", i+1),
			Role:     role,
			Status:   "active",
		}
	}

	return inputs
}

// UserWithRole returns a user with a specific role
func (u *UserTestData) UserWithRole(role string) model.User {
	now := time.Now()
	hashedPassword, _ := hash.HashPassword("TestPassword@123")

	return model.User{
		ID:        1,
		Name:      fmt.Sprintf("%s User", role),
		Email:     fmt.Sprintf("%s@example.com", role),
		Password:  hashedPassword,
		Role:      role,
		Status:    "active",
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// UserWithStatus returns a user with a specific status
func (u *UserTestData) UserWithStatus(status string) model.User {
	now := time.Now()
	hashedPassword, _ := hash.HashPassword("TestPassword@123")

	return model.User{
		ID:        1,
		Name:      "Test User",
		Email:     "test@example.com",
		Password:  hashedPassword,
		Role:      "user",
		Status:    status,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
