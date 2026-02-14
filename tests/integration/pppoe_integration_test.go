//go:build integration
// +build integration

package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	. "github.com/smartystreets/goconvey/convey"
	"gorm.io/gorm"

	postgres_outbound_adapter "go-template/internal/adapter/outbound/postgres"
	"go-template/internal/model"
	"go-template/tests/helpers"
)

func TestPppoeIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Use shared helper for container setup
	pgContainer, err := helpers.SetupPostgresContainer(ctx)
	if err != nil {
		t.Fatalf("Failed to setup postgres container: %v", err)
	}
	defer pgContainer.Terminate(ctx)

	// Use GORM AutoMigrate for all PPPoE related tables
	err = pgContainer.DB.AutoMigrate(
		&model.PppoeSecret{},
		&model.PppoeProfile{},
		&model.MikrotikRouter{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate pppoe tables: %v", err)
	}

	mikrotikAdapter := postgres_outbound_adapter.NewMikrotikAdapter(pgContainer.DB)

	Convey("Test PPPoE Integration with PostgreSQL", t, func() {
		// Setup Mikrotik Router with unique UUID
		routerID := uuid.New()
		router := &model.MikrotikRouter{
			ID:       routerID,
			Name:     "Integration Test Router",
			Address:  "192.168.88.1:8728",
			Username: "admin",
			Password: "admin",
			IsActive: &[]bool{true}[0],
		}

		// Cleanup before test
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.PppoeSecret{})
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.PppoeProfile{})
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.MikrotikRouter{})
		defer func() {
			pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.PppoeSecret{})
			pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.PppoeProfile{})
			pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.MikrotikRouter{})
		}()

		// Create router before all tests
		err := mikrotikAdapter.Create(router)
		So(err, ShouldBeNil)

		Convey("Router Management", func() {
			Convey("GetProfile retrieves router by ID", func() {
				found, err := mikrotikAdapter.FindByID(routerID.String())
				So(err, ShouldBeNil)
				So(found.Name, ShouldEqual, router.Name)
			})

			Convey("Update router updates fields", func() {
				router.Name = "Updated Router Name"
				err := mikrotikAdapter.Update(router)
				So(err, ShouldBeNil)

				found, err := mikrotikAdapter.FindByID(routerID.String())
				So(err, ShouldBeNil)
				So(found.Name, ShouldEqual, "Updated Router Name")
			})

			Convey("FindAll returns all routers", func() {
				router2 := &model.MikrotikRouter{
					ID:       uuid.MustParse("00000000-0000-0000-0000-000000000002"),
					Name:     "Second Router",
					Address:  "192.168.88.2:8728",
					Username: "admin",
					Password: "admin",
					IsActive: &[]bool{true}[0],
				}
				err := mikrotikAdapter.Create(router2)
				So(err, ShouldBeNil)

				routers, err := mikrotikAdapter.FindAll()
				So(err, ShouldBeNil)
				So(len(routers), ShouldBeGreaterThanOrEqualTo, 2)

				// Cleanup
				pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(router2)
			})

			Convey("Delete removes router", func() {
				router3 := &model.MikrotikRouter{
					ID:       uuid.MustParse("00000000-0000-0000-0000-000000000003"),
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

		Convey("Router Field Management", func() {
			Convey("Update router with additional fields", func() {
				router4 := &model.MikrotikRouter{
					ID:       uuid.MustParse("00000000-0000-0000-0000-000000000004"),
					Name:     "Secret Test Router",
					Address:  "192.168.88.4:8728",
					Username: "admin",
					Password: "admin",
					IsActive: &[]bool{true}[0],
				}
				err := mikrotikAdapter.Create(router4)
				So(err, ShouldBeNil)

				// Update router with additional fields
				router4.ApiPort = &[]int{8728}[0]
				router4.RestPort = &[]int{80}[0]
				router4.UseSSL = &[]bool{false}[0]
				router4.RouterOSVersion = &[]string{"7.1"}[0]
				router4.Identity = &[]string{"mikrotik-test"}[0]

				err = mikrotikAdapter.Update(router4)
				So(err, ShouldBeNil)

				updated, err := mikrotikAdapter.FindByID(router4.ID.String())
				So(err, ShouldBeNil)
				So(*updated.ApiPort, ShouldEqual, 8728)
				So(*updated.RestPort, ShouldEqual, 80)
				So(*updated.UseSSL, ShouldEqual, false)
				So(*updated.RouterOSVersion, ShouldEqual, "7.1")
				So(*updated.Identity, ShouldEqual, "mikrotik-test")

				// Cleanup
				pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(router4)
			})

			Convey("Test IsActive and LastSeenAt fields", func() {
				router5 := &model.MikrotikRouter{
					ID:       uuid.MustParse("00000000-0000-0000-0000-000000000005"),
					Name:     "Field Test Router",
					Address:  "192.168.88.5:8728",
					Username: "admin",
					Password: "admin",
					IsActive: &[]bool{true}[0],
				}
				err := mikrotikAdapter.Create(router5)
				So(err, ShouldBeNil)

				router5.IsActive = &[]bool{false}[0]
				now := time.Date(2026, time.February, 13, 12, 0, 0, 0, time.UTC)
				router5.LastSeenAt = &now
				err = mikrotikAdapter.Update(router5)
				So(err, ShouldBeNil)

				updated, err := mikrotikAdapter.FindByID(router5.ID.String())
				So(err, ShouldBeNil)
				So(*updated.IsActive, ShouldEqual, false)
				So(updated.LastSeenAt, ShouldNotBeNil)

				// Cleanup
				pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(router5)
			})
		})

		Convey("Error Cases", func() {
			Convey("FindByID with non-existent UUID returns error", func() {
				nonExistentID := uuid.MustParse("00000000-0000-0000-0000-999999999999")
				_, err := mikrotikAdapter.FindByID(nonExistentID.String())
				So(err, ShouldNotBeNil)
			})

			Convey("Delete with non-existent UUID does not error", func() {
				nonExistentID := uuid.MustParse("00000000-0000-0000-0000-888888888888")
				err := mikrotikAdapter.Delete(nonExistentID.String())
				So(err, ShouldBeNil)
			})
		})

		Convey("Password encryption field works", func() {
			router6 := &model.MikrotikRouter{
				ID:                uuid.MustParse("00000000-0000-0000-0000-000000000006"),
				Name:              "Password Test Router",
				Address:           "192.168.88.6:8728",
				Username:          "admin",
				Password:          "plain-password",
				PasswordEncrypted: &[]string{"encrypted-password"}[0],
				UseSSL:            &[]bool{true}[0],
				IsActive:          &[]bool{true}[0],
			}
			err := mikrotikAdapter.Create(router6)
			So(err, ShouldBeNil)

			retrieved, err := mikrotikAdapter.FindByID(router6.ID.String())
			So(err, ShouldBeNil)
			So(retrieved.Password, ShouldEqual, "plain-password")
			So(retrieved.PasswordEncrypted, ShouldNotBeNil)
			So(*retrieved.PasswordEncrypted, ShouldEqual, "encrypted-password")
			So(*retrieved.UseSSL, ShouldEqual, true)

			// Cleanup
			pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(router6)
		})
	})
}
