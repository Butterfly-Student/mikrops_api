package postgres_outbound_adapter_test

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"gorm.io/gorm"

	postgres_outbound_adapter "go-template/internal/adapter/outbound/postgres"
	"go-template/internal/model"
	"go-template/tests/helpers"
)

func TestClientAdapter(t *testing.T) {
	// Integration tests usually take longer, skip in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Start Postgres Container
	pgContainer, err := helpers.SetupPostgresContainer(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer pgContainer.Terminate(ctx)

	// AutoMigrate the schema
	err = pgContainer.DB.AutoMigrate(&model.Client{})
	if err != nil {
		t.Fatal(err)
	}

	adapter := postgres_outbound_adapter.NewClientAdapter(pgContainer.DB)

	Convey("Test Postgres Client Adapter (Integration)", t, func() {
		// Cleanup before each test to ensure clean state
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.Client{})

		input := model.ClientInput{
			FullName: "Test Admin User",
			Email:    "test@example.com",
			Password: "password123",
			Role:     "admin",
		}

		Convey("Upsert", func() {
			Convey("Insert new record", func() {
				err := adapter.Upsert([]model.ClientInput{input})
				So(err, ShouldBeNil)

				var count int64
				pgContainer.DB.Model(&model.Client{}).Count(&count)
				So(count, ShouldEqual, 1)

				var stored model.Client
				pgContainer.DB.First(&stored)
				So(stored.FullName, ShouldEqual, input.FullName)
				So(stored.Email, ShouldEqual, input.Email)
			})

			Convey("Update existing record (Conflict on Email)", func() {
				// First insert
				adapter.Upsert([]model.ClientInput{input})

				// Update data
				updatedInput := input
				updatedInput.FullName = "Updated Name"

				// Same Email -> Should Update
				err := adapter.Upsert([]model.ClientInput{updatedInput})
				So(err, ShouldBeNil)

				var stored model.Client
				pgContainer.DB.First(&stored, "email = ?", input.Email)
				So(stored.FullName, ShouldEqual, "Updated Name")

				var count int64
				pgContainer.DB.Model(&model.Client{}).Count(&count)
				So(count, ShouldEqual, 1)
			})
		})

		Convey("FindByFilter", func() {
			// Seed data
			adapter.Upsert([]model.ClientInput{input})

			// Get actual record
			var stored model.Client
			pgContainer.DB.First(&stored, "email = ?", input.Email)

			Convey("Find by Email", func() {
				filter := model.ClientFilter{Emails: []string{input.Email}}
				results, err := adapter.FindByFilter(filter, false)
				So(err, ShouldBeNil)
				So(len(results), ShouldEqual, 1)
				So(results[0].Email, ShouldEqual, input.Email)
			})

			Convey("Find by Role", func() {
				filter := model.ClientFilter{Roles: []model.AdminUserRole{model.AdminRoleAdmin}}
				results, err := adapter.FindByFilter(filter, false)
				So(err, ShouldBeNil)
				So(len(results), ShouldEqual, 1)
				So(results[0].Role, ShouldEqual, model.AdminRoleAdmin)
			})

			Convey("With Lock", func() {
				filter := model.ClientFilter{Emails: []string{input.Email}}
				results, err := adapter.FindByFilter(filter, true)
				So(err, ShouldBeNil)
				So(len(results), ShouldEqual, 1)
			})

			Convey("Empty Result", func() {
				filter := model.ClientFilter{Emails: []string{"nonexistent@example.com"}}
				results, err := adapter.FindByFilter(filter, false)
				So(err, ShouldBeNil)
				So(len(results), ShouldEqual, 0)
			})
		})

		Convey("IsExists", func() {
			adapter.Upsert([]model.ClientInput{input})

			Convey("Exists", func() {
				exists, err := adapter.IsExists(input.Email)
				So(err, ShouldBeNil)
				So(exists, ShouldBeTrue)
			})

			Convey("Not Exists", func() {
				exists, err := adapter.IsExists("nonexistent@example.com")
				So(err, ShouldBeNil)
				So(exists, ShouldBeFalse)
			})
		})

		Convey("DeleteByFilter", func() {
			adapter.Upsert([]model.ClientInput{input})

			Convey("Delete by Email", func() {
				filter := model.ClientFilter{Emails: []string{input.Email}}
				err := adapter.DeleteByFilter(filter)
				So(err, ShouldBeNil)

				var count int64
				pgContainer.DB.Model(&model.Client{}).Count(&count)
				So(count, ShouldEqual, 0)
			})
		})
	})
}
