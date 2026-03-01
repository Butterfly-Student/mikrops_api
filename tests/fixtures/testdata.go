package fixtures

import (
	"go-template/internal/model"

	"github.com/google/uuid"
)

type ClientTestData struct{}

func NewClientTestData() *ClientTestData {
	return &ClientTestData{}
}

func (c *ClientTestData) ValidClientInput() model.AdminUserInput {
	return model.AdminUserInput{
		FullName: "Test Client",
		Email:    "test-client@example.com",
		Password: "Test@123",
		Role:     string(model.AdminRoleAdmin),
	}
}

func (c *ClientTestData) ValidClient() model.AdminUser {
	isActive := true
	return model.AdminUser{
		ID:       uuid.New(),
		FullName: "Test Client",
		Email:    "test-client@example.com",
		Role:     model.AdminRoleAdmin,
		IsActive: &isActive,
	}
}

func (c *ClientTestData) ValidClientFilter() model.AdminUserFilter {
	return model.AdminUserFilter{
		Emails:     []string{"test-client@example.com"},
		BearerKeys: []string{"test-bearer-key"},
	}
}

func (c *ClientTestData) MultipleClients(count int) []model.AdminUser {
	clients := make([]model.AdminUser, count)
	isActive := true
	for i := 0; i < count; i++ {
		clients[i] = model.AdminUser{
			ID:       uuid.New(),
			FullName: "Client " + string(rune('A'+i)),
			Email:    "client-" + string(rune('a'+i)) + "@example.com",
			Role:     model.AdminRoleCS,
			IsActive: &isActive,
		}
	}
	return clients
}

func (c *ClientTestData) MultipleClientInputs(count int) []model.AdminUserInput {
	inputs := make([]model.AdminUserInput, count)
	for i := 0; i < count; i++ {
		inputs[i] = model.AdminUserInput{
			FullName: "Client " + string(rune('A'+i)),
			Email:    "client-" + string(rune('a'+i)) + "@example.com",
			Password: "Test@123",
			Role:     string(model.AdminRoleCS),
		}
	}
	return inputs
}
