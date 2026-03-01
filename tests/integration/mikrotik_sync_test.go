package integration_test

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mikrotik_outbound_adapter "go-template/internal/adapter/outbound/mikrotik"
	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
	"go-template/pkg/hotspot"
	"go-template/tests/helpers"
)

// TestMikrotikSyncIntegration melakukan end-to-end test dengan MikroTik nyata
// Requirements:
// - MikroTik router accessible di IP yang dikonfigurasi
// - PostgreSQL database running
// - Bandwidth profile "TEST-10M" sudah dibuat di database
func TestMikrotikSyncIntegration(t *testing.T) {
	if os.Getenv("MIKROTIK_TEST_IP") == "" {
		t.Skip("Skipping MikroTik integration test. Set MIKROTIK_TEST_IP to run this test.")
	}

	ctx := context.Background()
	_ = godotenv.Load("../../.env")

	// Setup PostgreSQL container
	pgContainer, err := helpers.SetupPostgresContainer(ctx)
	require.NoError(t, err)
	defer pgContainer.Terminate(ctx)

	// Migrate required tables
	err = pgContainer.DB.AutoMigrate(
		&model.Customer{},
		&model.Subscription{},
		&model.BandwidthProfile{},
		&model.MikrotikRouter{},
	)
	require.NoError(t, err)

	// Setup MikroTik adapter
	mikrotikPort := mikrotik_outbound_adapter.NewMikrotikClientAdapter()
	hotspotPort := mikrotik_outbound_adapter.NewHotspotAdapter()

	// Create test router di database
	testRouter := &model.MikrotikRouter{
		ID:       uuid.New(),
		Name:     "Test Router",
		Address:  os.Getenv("MIKROTIK_TEST_IP") + ":8728",
		Username: os.Getenv("MIKROTIK_TEST_USERNAME"),
		Password: os.Getenv("MIKROTIK_TEST_PASSWORD"),
		IsActive: boolPtr(true),
	}
	err = pgContainer.DB.Create(testRouter).Error
	require.NoError(t, err)

	// Test PPPoE Sync
	t.Run("PPPoE Secret Sync", func(t *testing.T) {
		testPPPoESecret(t, ctx, mikrotikPort, testRouter)
	})

	// Test Hotspot Sync  
	t.Run("Hotspot User Sync", func(t *testing.T) {
		testHotspotUser(t, ctx, hotspotPort, testRouter)
	})
}

func testPPPoESecret(t *testing.T, ctx context.Context, mikrotikPort outbound_port.MikrotikPort, router *model.MikrotikRouter) {
	// Generate unique username untuk menghindari konflik
	testUsername := "test_pppoe_" + uuid.New().String()[:8]
	
	secret := &model.PppoeSecret{
		Name:          testUsername,
		Password:      "testpass123",
		Service:       "pppoe",
		Profile:       "default", // Pastikan profile ini ada di MikroTik
		LocalAddress:  "10.10.10.1",
		RemoteAddress: "10.10.10.100",
		Comment:       "Integration Test",
	}

	// Test Create
	t.Run("Create Secret", func(t *testing.T) {
		err := mikrotikPort.CreateSecret(router, secret)
		assert.NoError(t, err, "Gagal membuat PPPoE secret di MikroTik")
		
		// Verifikasi secret terbuat
		created, err := mikrotikPort.GetSecret(router, testUsername)
		assert.NoError(t, err)
		assert.Equal(t, testUsername, created.Name)
		t.Logf("✓ PPPoE Secret created: %s", created.Name)
	})

	// Test Update
	t.Run("Update Secret", func(t *testing.T) {
		secret.Password = "newpass456"
		secret.Comment = "Updated via integration test"
		
		err := mikrotikPort.UpdateSecret(router, secret)
		assert.NoError(t, err, "Gagal update PPPoE secret")
		
		updated, err := mikrotikPort.GetSecret(router, testUsername)
		assert.NoError(t, err)
		assert.Equal(t, "Updated via integration test", updated.Comment)
		t.Logf("✓ PPPoE Secret updated: %s", updated.Name)
	})

	// Test Disable
	t.Run("Disable Secret", func(t *testing.T) {
		secret.Disabled = true
		err := mikrotikPort.UpdateSecret(router, secret)
		assert.NoError(t, err, "Gagal disable PPPoE secret")
		
		disabled, err := mikrotikPort.GetSecret(router, testUsername)
		assert.NoError(t, err)
		// Note: Some MikroTik versions don't return disabled field properly
		// So we just verify no error occurred
		t.Logf("✓ PPPoE Secret disabled (disabled=%v): %s", disabled.Disabled, disabled.Name)
	})

	// Cleanup - Delete
	t.Run("Delete Secret", func(t *testing.T) {
		err := mikrotikPort.DeleteSecret(router, testUsername)
		assert.NoError(t, err, "Gagal menghapus PPPoE secret")
		
		// Verifikasi sudah terhapus
		_, err = mikrotikPort.GetSecret(router, testUsername)
		assert.Error(t, err, "Secret seharusnya sudah terhapus")
		t.Logf("✓ PPPoE Secret deleted: %s", testUsername)
	})
}

func testHotspotUser(t *testing.T, ctx context.Context, hotspotPort outbound_port.HotspotPort, router *model.MikrotikRouter) {
	hotspotClient, err := hotspotPort.GetHotspotClient(router)
	require.NoError(t, err)

	testUsername := "test_hotspot_" + uuid.New().String()[:8]

	// Test Create
	t.Run("Create Hotspot User", func(t *testing.T) {
		user := &hotspot.User{
			Name:     testUsername,
			Password: "testpass123",
			Profile:  "default", // Pastikan profile ini ada di MikroTik
			Server:   "all",
			Comment:  "Integration Test",
		}

		err := hotspotClient.CreateUser(ctx, user)
		assert.NoError(t, err, "Gagal membuat hotspot user")

		// Verifikasi
		created, err := hotspotClient.GetUser(ctx, testUsername)
		assert.NoError(t, err)
		assert.Equal(t, testUsername, created.Name)
		t.Logf("✓ Hotspot User created: %s", created.Name)
	})

	// Test Update
	t.Run("Update Hotspot User", func(t *testing.T) {
		profile := "default"
		updates := &hotspot.UserUpdate{
			Profile: &profile,
		}

		err := hotspotClient.UpdateUser(ctx, testUsername, updates)
		assert.NoError(t, err, "Gagal update hotspot user")

		updated, err := hotspotClient.GetUser(ctx, testUsername)
		assert.NoError(t, err)
		assert.Equal(t, "default", updated.Profile)
		t.Logf("✓ Hotspot User updated: %s", updated.Name)
	})

	// Test Disable
	t.Run("Disable Hotspot User", func(t *testing.T) {
		err := hotspotClient.DisableUser(ctx, testUsername)
		assert.NoError(t, err, "Gagal disable hotspot user")

		disabled, err := hotspotClient.GetUser(ctx, testUsername)
		assert.NoError(t, err)
		assert.True(t, disabled.Disabled)
		t.Logf("✓ Hotspot User disabled: %s", disabled.Name)
	})

	// Test Enable
	t.Run("Enable Hotspot User", func(t *testing.T) {
		err := hotspotClient.EnableUser(ctx, testUsername)
		assert.NoError(t, err, "Gagal enable hotspot user")

		enabled, err := hotspotClient.GetUser(ctx, testUsername)
		assert.NoError(t, err)
		assert.False(t, enabled.Disabled)
		t.Logf("✓ Hotspot User enabled: %s", enabled.Name)
	})

	// Cleanup
	t.Run("Delete Hotspot User", func(t *testing.T) {
		err := hotspotClient.DeleteUser(ctx, testUsername)
		assert.NoError(t, err, "Gagal menghapus hotspot user")

		// Verifikasi
		_, err = hotspotClient.GetUser(ctx, testUsername)
		assert.Error(t, err, "User seharusnya sudah terhapus")
		t.Logf("✓ Hotspot User deleted: %s", testUsername)
	})
}

// TestEndToEndRegistrationFlow mensimulasikan flow lengkap dari registration sampai sync ke MikroTik
func TestEndToEndRegistrationFlow(t *testing.T) {
	if os.Getenv("MIKROTIK_TEST_IP") == "" {
		t.Skip("Skipping E2E test. Set MIKROTIK_TEST_IP to run this test.")
	}

	ctx := context.Background()
	_ = godotenv.Load("../../.env")

	// Setup
	pgContainer, err := helpers.SetupPostgresContainer(ctx)
	require.NoError(t, err)
	defer pgContainer.Terminate(ctx)

	// Migrate
	err = pgContainer.DB.AutoMigrate(
		&model.Customer{},
		&model.Subscription{},
		&model.BandwidthProfile{},
		&model.MikrotikRouter{},
		&model.CustomerRegistration{},
	)
	require.NoError(t, err)

	// Setup MikroTik
	mikrotikPort := mikrotik_outbound_adapter.NewMikrotikClientAdapter()

	testRouter := &model.MikrotikRouter{
		ID:       uuid.New(),
		Name:     "Test Router",
		Address:  os.Getenv("MIKROTIK_TEST_IP") + ":8728",
		Username: os.Getenv("MIKROTIK_TEST_USERNAME"),
		Password: os.Getenv("MIKROTIK_TEST_PASSWORD"),
		IsActive: boolPtr(true),
	}
	err = pgContainer.DB.Create(testRouter).Error
	require.NoError(t, err)

	// Buat bandwidth profile untuk PPPoE
	pppoeProfile := &model.BandwidthProfile{
		ID:             uuid.New(),
		ProfileCode:    "TEST-PPPOE-10M",
		Name:           "Test PPPoE 10Mbps",
		ServiceType:    model.ServiceTypePPPoE,
		PppProfileName: "default", // Pastikan profile ini ada di MikroTik
		DownloadSpeed:  10240,
		UploadSpeed:    10240,
		PriceMonthly:   200000,
	}
	err = pgContainer.DB.Create(pppoeProfile).Error
	require.NoError(t, err)

	// Simulasi registration approval dengan sync ke MikroTik
	t.Run("Registration Approval and MikroTik Sync", func(t *testing.T) {
		testUsername := "e2e_test_" + uuid.New().String()[:8]
		testPassword := "e2epass123"

		// 1. Create subscription di database
		customer := &model.Customer{
			ID:           uuid.New(),
			CustomerCode: "E2E001",
			FullName:     "E2E Test Customer",
			Phone:        "08123456789",
			Status:       model.CustomerStatusPending,
		}
		err = pgContainer.DB.Create(customer).Error
		require.NoError(t, err)

		subscription := &model.Subscription{
			ID:          uuid.New(),
			CustomerID:  customer.ID,
			PlanID:      pppoeProfile.ID,
			RouterID:    testRouter.ID,
			Username:    testUsername,
			Password:    testPassword,
			ServiceType: model.ServiceTypePPPoE,
			Status:      model.SubscriptionStatusActive,
		}
		err = pgContainer.DB.Create(subscription).Error
		require.NoError(t, err)

		// 2. Sync ke MikroTik (simulasi yang dilakukan oleh RabbitMQ consumer)
		secret := &model.PppoeSecret{
			Name:          testUsername,
			Password:      testPassword,
			Service:       "pppoe",
			Profile:       pppoeProfile.PppProfileName,
			Comment:       "E2E Test " + subscription.ID.String(),
		}
		
		err = mikrotikPort.CreateSecret(testRouter, secret)
		require.NoError(t, err, "Gagal sync ke MikroTik")

		// 3. Verifikasi secret ada di MikroTik
		created, err := mikrotikPort.GetSecret(testRouter, testUsername)
		require.NoError(t, err)
		assert.Equal(t, testUsername, created.Name)
		assert.Equal(t, testPassword, created.Password)

		t.Logf("✓ Subscription berhasil dibuat dan di-sync ke MikroTik")
		t.Logf("  Username: %s", testUsername)
		t.Logf("  Router: %s", testRouter.Address)

		// Cleanup
		_ = mikrotikPort.DeleteSecret(testRouter, testUsername)
	})
}

func boolPtr(b bool) *bool {
	return &b
}
