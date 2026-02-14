//go:build integration
// +build integration

package integration_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	. "github.com/smartystreets/goconvey/convey"
	"gorm.io/gorm"

	postgres_outbound_adapter "go-template/internal/adapter/outbound/postgres"
	"go-template/internal/model"
	"go-template/tests/helpers"
	"go-template/utils/hash"
)

func TestMikrotikIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	os.Setenv("JWT_SECRET", "test-secret-key")
	os.Setenv("JWT_REFRESH_SECRET", "test-refresh-secret-key")
	defer os.Unsetenv("JWT_SECRET")
	defer os.Unsetenv("JWT_REFRESH_SECRET")

	ctx := context.Background()

	pgContainer, err := helpers.SetupPostgresContainer(ctx)
	if err != nil {
		t.Fatalf("Failed to setup postgres container: %v", err)
	}
	defer pgContainer.Terminate(ctx)

	err = pgContainer.DB.AutoMigrate(
		&model.User{},
		&model.MikrotikRouter{},
		&model.PppoeSecret{},
		&model.PppoeProfile{},
		&model.PppoeQueue{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate mikrotik tables: %v", err)
	}

	userAdapter := postgres_outbound_adapter.NewUserAdapter(pgContainer.DB)
	mikrotikAdapter := postgres_outbound_adapter.NewMikrotikAdapter(pgContainer.DB)

	Convey("Test Mikrotik Integration with PostgreSQL", t, func() {
		// Cleanup before test
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.PppoeQueue{})
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.PppoeProfile{})
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.PppoeSecret{})
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.MikrotikRouter{})
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.User{})

		// Setup User
		email := "mikrotik-integration-" + time.Now().Format("20060102150405") + "@example.com"
		password := "TestPassword123!"
		hashedPassword, _ := hash.HashPassword(password)
		user := &model.User{
			Name:     "Mikrotik Integration User",
			Email:    email,
			Password: hashedPassword,
			Role:     "admin",
			Status:   "active",
		}
		err := userAdapter.Create(user)
		So(err, ShouldBeNil)

		Convey("Router CRUD Operations", func() {
			routerID := uuid.New()
			router := &model.MikrotikRouter{
				ID:       routerID,
				Name:     "Integration Mikrotik Router",
				Address:  "192.168.88.1:8728",
				Username: "admin",
				Password: "admin",
				IsActive: &[]bool{true}[0],
			}

			err := mikrotikAdapter.Create(router)
			So(err, ShouldBeNil)

			Convey("FindByID retrieves router", func() {
				found, err := mikrotikAdapter.FindByID(routerID.String())
				So(err, ShouldBeNil)
				So(found.ID, ShouldEqual, routerID)
				So(found.Name, ShouldEqual, "Integration Mikrotik Router")
			})

			Convey("Update router", func() {
				router.Name = "Updated Mikrotik Router"
				err := mikrotikAdapter.Update(router)
				So(err, ShouldBeNil)

				found, err := mikrotikAdapter.FindByID(routerID.String())
				So(err, ShouldBeNil)
				So(found.Name, ShouldEqual, "Updated Mikrotik Router")
			})

			Convey("FindAll returns all routers", func() {
				router2 := &model.MikrotikRouter{
					ID:       uuid.New(),
					Name:     "Second Mikrotik Router",
					Address:  "192.168.88.2:8728",
					Username: "admin",
					Password: "admin",
					IsActive: &[]bool{true}[0],
				}
				err := mikrotikAdapter.Create(router2)
				So(err, ShouldBeNil)

				allRouters, err := mikrotikAdapter.FindAll()
				So(err, ShouldBeNil)
				So(len(allRouters), ShouldBeGreaterThanOrEqualTo, 2)

				// Cleanup
				pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(router2)
			})

			Convey("Delete removes router", func() {
				router3 := &model.MikrotikRouter{
					ID:       uuid.New(),
					Name:     "Router To Delete",
					Address:  "192.168.88.3:8728",
					Username: "admin",
					Password: "admin",
					IsActive: &[]bool{true}[0],
				}
				err := mikrotikAdapter.Create(router3)
				So(err, ShouldBeNil)

				err = mikrotikAdapter.Delete(router3.ID.String())
				So(err, ShouldBeNil)

				_, err = mikrotikAdapter.FindByID(router3.ID.String())
				So(err, ShouldNotBeNil)
			})
		})
	})
}
