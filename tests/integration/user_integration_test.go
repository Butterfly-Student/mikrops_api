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

	postgres_outbound_adapter "go-template/internal/adapter/outbound/postgres"
	"go-template/internal/model"
	"go-template/tests/helpers"
)

func TestUserIntegration(t *testing.T) {
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

	// Use GORM AutoMigrate
	err = pgContainer.DB.AutoMigrate(&model.User{})
	if err != nil {
		t.Fatalf("Failed to migrate table: %v", err)
	}

	Convey("Test User Integration with PostgreSQL", t, func() {
		adapter := postgres_outbound_adapter.NewUserAdapter(pgContainer.DB)

		// Cleanup before test
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.User{})

		Convey("Full CRUD cycle", func() {
			email := "integration-" + time.Now().Format("20060102150405") + "@example.com"
			user := model.User{
				Name:     "Integration User",
				Email:    email,
				Password: "hashedpassword",
				Role:     "user",
				Status:   "active",
			}

			Convey("Create creates a new user", func() {
				err := adapter.Create(&user)
				So(err, ShouldBeNil)
				So(user.ID, ShouldBeGreaterThan, 0)

				Convey("FindByID retrieves the user", func() {
					found, err := adapter.FindByID(user.ID)
					So(err, ShouldBeNil)
					So(found.Email, ShouldEqual, email)
				})

				Convey("FindByEmail retrieves the user", func() {
					found, err := adapter.FindByEmail(email)
					So(err, ShouldBeNil)
					So(found.ID, ShouldEqual, user.ID)
				})

				Convey("Update updates the user", func() {
					user.Name = "Updated Name"
					err := adapter.Update(user)
					So(err, ShouldBeNil)

					found, err := adapter.FindByID(user.ID)
					So(err, ShouldBeNil)
					So(found.Name, ShouldEqual, "Updated Name")
				})
			})
		})

		Convey("FindByEmail returns error for non-existent user", func() {
			_, err := adapter.FindByEmail("nonexistent@example.com")
			So(err, ShouldNotBeNil)
		})

		Convey("FindByID returns error for non-existent user", func() {
			_, err := adapter.FindByID(99999)
			So(err, ShouldNotBeNil)
		})
	})
}
