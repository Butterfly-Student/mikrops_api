//go:build integration
// +build integration

package integration_test

import (
	"context"
	"testing"
	"time"

	_ "github.com/lib/pq"
	. "github.com/smartystreets/goconvey/convey"
	"gorm.io/gorm"

	"go-template/internal/model"
	"go-template/tests/helpers"
)

func TestRegistrationIntegration(t *testing.T) {
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

	// Use GORM AutoMigrate for all related tables
	err = pgContainer.DB.AutoMigrate(
		&model.CustomerRegistration{},
		&model.BandwidthProfile{},
		&model.MikrotikRouter{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate tables: %v", err)
	}

	Convey("Test Registration Integration with PostgreSQL", t, func() {
		// Cleanup before test
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.CustomerRegistration{})

		Convey("Full registration lifecycle", func() {
			Convey("Create a new registration", func() {
				registration := model.CustomerRegistration{
					FullName: "Test Registration",
					Email:    strPtr("registration@example.com"),
					Phone:    "08123456789",
					Address:  strPtr("Test Address"),
					Status:   model.RegistrationStatusPending,
				}

				err := pgContainer.DB.Create(&registration).Error
				So(err, ShouldBeNil)
				So(registration.ID.String(), ShouldNotBeEmpty)
				So(registration.Status, ShouldEqual, model.RegistrationStatusPending)

				Convey("Retrieve registration by ID", func() {
					var found model.CustomerRegistration
					err := pgContainer.DB.First(&found, "id = ?", registration.ID).Error
					So(err, ShouldBeNil)
					So(found.FullName, ShouldEqual, "Test Registration")
				})

				Convey("Update registration status", func() {
					registration.Status = model.RegistrationStatusApproved
					err := pgContainer.DB.Save(&registration).Error
					So(err, ShouldBeNil)

					var updated model.CustomerRegistration
					err = pgContainer.DB.First(&updated, "id = ?", registration.ID).Error
					So(err, ShouldBeNil)
					So(updated.Status, ShouldEqual, model.RegistrationStatusApproved)
				})

				Convey("Soft delete registration", func() {
					err := pgContainer.DB.Delete(&registration).Error
					So(err, ShouldBeNil)

					var found model.CustomerRegistration
					err = pgContainer.DB.First(&found, "id = ?", registration.ID).Error
					So(err, ShouldNotBeNil) // Should not be found after soft delete
				})
			})

			Convey("List registrations with filter", func() {
				// Create multiple registrations
				for i := 0; i < 3; i++ {
					reg := model.CustomerRegistration{
						FullName: "List Registration " + string(rune('A'+i)),
						Phone:    "0812345678" + string(rune('0'+i)),
						Status:   model.RegistrationStatusPending,
					}
					pgContainer.DB.Create(&reg)
				}

				var registrations []model.CustomerRegistration
				err := pgContainer.DB.Where("status = ?", model.RegistrationStatusPending).Find(&registrations).Error
				So(err, ShouldBeNil)
				So(len(registrations), ShouldBeGreaterThanOrEqualTo, 3)
			})
		})
	})
}
