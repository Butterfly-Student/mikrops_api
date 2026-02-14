package fixtures

import (
	"fmt"
	"time"

	"go-template/internal/model"

	"github.com/google/uuid"
)

type MikrotikTestData struct{}

func NewMikrotikTestData() *MikrotikTestData {
	return &MikrotikTestData{}
}

// ValidMikrotikRouter returns a valid MikroTik router for testing
func (m *MikrotikTestData) ValidMikrotikRouter() model.MikrotikRouter {
	now := time.Now()
	apiPort := 8728
	restPort := 80
	useSSL := false
	isActive := true
	routerOSVersion := "7.10"
	identity := "MikroTik-Test"

	return model.MikrotikRouter{
		ID:              uuid.New(),
		Name:            "Test Router",
		Address:         "192.168.1.1:8728",
		ApiPort:         &apiPort,
		RestPort:        &restPort,
		Username:        "admin",
		Password:        "test-password",
		UseSSL:          &useSSL,
		RouterOSVersion: &routerOSVersion,
		Identity:        &identity,
		IsActive:        &isActive,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

// ValidMikrotikRouterInput returns a valid MikroTik router input for testing
func (m *MikrotikTestData) ValidMikrotikRouterInput() model.MikrotikRouter {
	router := m.ValidMikrotikRouter()
	return router
}

// ActiveRouter returns an active MikroTik router
func (m *MikrotikTestData) ActiveRouter() model.MikrotikRouter {
	router := m.ValidMikrotikRouter()
	isActive := true
	router.IsActive = &isActive
	return router
}

// InactiveRouter returns an inactive MikroTik router
func (m *MikrotikTestData) InactiveRouter() model.MikrotikRouter {
	router := m.ValidMikrotikRouter()
	isActive := false
	router.IsActive = &isActive
	return router
}

// RouterWithSSL returns a MikroTik router configured with SSL
func (m *MikrotikTestData) RouterWithSSL() model.MikrotikRouter {
	router := m.ValidMikrotikRouter()
	useSSL := true
	router.UseSSL = &useSSL
	router.ApiPort = intPtr(8729)
	return router
}

// MultipleRouters returns multiple MikroTik routers for testing
func (m *MikrotikTestData) MultipleRouters(count int) []model.MikrotikRouter {
	routers := make([]model.MikrotikRouter, count)
	now := time.Now()
	apiPort := 8728
	restPort := 80
	useSSL := false

	for i := 0; i < count; i++ {
		isActive := i%2 == 0 // Alternate between active and inactive
		routerOSVersion := "7.10"
		identity := fmt.Sprintf("MikroTik-Test-%d", i+1)

		routers[i] = model.MikrotikRouter{
			ID:              uuid.New(),
			Name:            fmt.Sprintf("Router %d", i+1),
			Address:         fmt.Sprintf("192.168.1.%d:8728", i+1),
			ApiPort:         &apiPort,
			RestPort:        &restPort,
			Username:        "admin",
			Password:        fmt.Sprintf("password-%d", i+1),
			UseSSL:          &useSSL,
			RouterOSVersion: &routerOSVersion,
			Identity:        &identity,
			IsActive:        &isActive,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
	}

	return routers
}

// RouterWithAddress returns a router with specific address
func (m *MikrotikTestData) RouterWithAddress(address string) model.MikrotikRouter {
	router := m.ValidMikrotikRouter()
	router.Address = address
	return router
}

// RouterWithName returns a router with specific name
func (m *MikrotikTestData) RouterWithName(name string) model.MikrotikRouter {
	router := m.ValidMikrotikRouter()
	router.Name = name
	return router
}

// Helper function to create int pointer
func intPtr(v int) *int {
	return &v
}

// Helper function to create bool pointer
func boolPtr(v bool) *bool {
	return &v
}

// Helper function to create string pointer
func stringPtr(v string) *string {
	return &v
}
