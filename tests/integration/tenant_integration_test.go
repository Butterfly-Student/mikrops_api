//go:build integration
// +build integration

package integration_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	_ "github.com/lib/pq"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"

	postgres_outbound_adapter "mikrops/internal/adapter/outbound/postgres"
	"mikrops/internal/model"
)

func TestTenantIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:14-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		t.Fatalf("Failed to start postgres container: %v", err)
	}
	defer pgContainer.Terminate(ctx)

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to get connection string: %v", err)
	}

	db, err := gorm.Open(gormpostgres.Open(connStr), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get sql.DB: %v", err)
	}
	defer sqlDB.Close()

	// Create tenants table
	_, err = sqlDB.Exec(`
		CREATE TABLE IF NOT EXISTS tenants (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			slug VARCHAR(100) UNIQUE NOT NULL,
			email VARCHAR(255),
			phone VARCHAR(50),
			address TEXT,
			logo_url VARCHAR(500),
			max_nas INTEGER DEFAULT 3,
			subscription_plan VARCHAR(50) DEFAULT 'basic',
			subscription_expires_at TIMESTAMP,
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
		);
		CREATE UNIQUE INDEX IF NOT EXISTS idx_tenants_slug ON tenants(slug);
	`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	Convey("Test Tenant Integration with PostgreSQL", t, func() {
		adapter := postgres_outbound_adapter.NewTenantAdapter(db)

		Convey("Full tenant CRUD cycle", func() {
			slug := fmt.Sprintf("integration-tenant-%d", time.Now().UnixNano())
			input := model.TenantInput{
				Name:             "Integration Test Tenant",
				Slug:             slug,
				Email:            "integration@test.com",
				Phone:            "+6281234567890",
				Address:          "Jl. Integration Test No. 123",
				MaxNas:           3,
				SubscriptionPlan: "basic",
				IsActive:         true,
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			}

			Convey("Create creates a new tenant", func() {
				tenant, err := adapter.Create(input)
				So(err, ShouldBeNil)
				So(tenant.ID, ShouldNotBeEmpty)
				So(tenant.Name, ShouldEqual, "Integration Test Tenant")
				So(tenant.Slug, ShouldEqual, slug)

				tenantID := tenant.ID

				Convey("FindByID retrieves the tenant", func() {
					result, err := adapter.FindByID(tenantID)
					So(err, ShouldBeNil)
					So(result.ID, ShouldEqual, tenantID)
					So(result.Name, ShouldEqual, "Integration Test Tenant")
				})

				Convey("FindByFilter retrieves the tenant", func() {
					filter := model.TenantFilter{Slugs: []string{slug}}
					results, err := adapter.FindByFilter(filter)
					So(err, ShouldBeNil)
					So(len(results), ShouldEqual, 1)
					So(results[0].Slug, ShouldEqual, slug)
				})

				Convey("Update modifies the tenant", func() {
					updateInput := input
					updateInput.Name = "Updated Tenant Name"
					updateInput.Email = "updated@test.com"

					err := adapter.Update(tenantID, updateInput)
					So(err, ShouldBeNil)

					result, err := adapter.FindByID(tenantID)
					So(err, ShouldBeNil)
					So(result.Name, ShouldEqual, "Updated Tenant Name")
					So(result.Email, ShouldEqual, "updated@test.com")
				})

				Convey("Delete removes the tenant", func() {
					err := adapter.Delete(tenantID)
					So(err, ShouldBeNil)

					_, err = adapter.FindByID(tenantID)
					So(err, ShouldNotBeNil) // Should not find deleted tenant
				})
			})
		})

		Convey("Unique slug constraint", func() {
			slug := fmt.Sprintf("unique-slug-%d", time.Now().UnixNano())
			input1 := model.TenantInput{
				Name:      "Tenant 1",
				Slug:      slug,
				Email:     "tenant1@test.com",
				IsActive:  true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			tenant1, err := adapter.Create(input1)
			So(err, ShouldBeNil)
			So(tenant1.ID, ShouldNotBeEmpty)

			input2 := model.TenantInput{
				Name:      "Tenant 2",
				Slug:      slug, // Same slug
				Email:     "tenant2@test.com",
				IsActive:  true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			_, err = adapter.Create(input2)
			So(err, ShouldNotBeNil) // Should fail due to unique constraint
		})

		Convey("Filter by active status", func() {
			now := time.Now()
			slugActive := fmt.Sprintf("active-tenant-%d", now.UnixNano())
			time.Sleep(1 * time.Millisecond) // Ensure different timestamp
			slugInactive := fmt.Sprintf("inactive-tenant-%d", time.Now().UnixNano())

			inputActive := model.TenantInput{
				Name:      "Active Tenant",
				Slug:      slugActive,
				Email:     "active@test.com",
				IsActive:  true,
				CreatedAt: now,
				UpdatedAt: now,
			}

			inputInactive := model.TenantInput{
				Name:      "Inactive Tenant",
				Slug:      slugInactive,
				Email:     "inactive@test.com",
				IsActive:  false,
				CreatedAt: now,
				UpdatedAt: now,
			}

			_, err := adapter.Create(inputActive)
			So(err, ShouldBeNil)

			_, err = adapter.Create(inputInactive)
			So(err, ShouldBeNil)

			isActive := true
			filter := model.TenantFilter{IsActive: &isActive}
			results, err := adapter.FindByFilter(filter)
			So(err, ShouldBeNil)

			// Should find at least the active tenant we just created
			activeFound := false
			for _, tenant := range results {
				if tenant.Slug == slugActive {
					activeFound = true
					So(tenant.IsActive, ShouldBeTrue)
				}
				So(tenant.IsActive, ShouldBeTrue) // All results should be active
			}
			So(activeFound, ShouldBeTrue)
		})
	})
}
