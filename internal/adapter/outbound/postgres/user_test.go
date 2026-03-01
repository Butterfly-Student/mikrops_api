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

func TestUserAdapter(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	pgContainer, err := helpers.SetupPostgresContainer(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer pgContainer.Terminate(ctx)

	err = pgContainer.DB.AutoMigrate(&model.AdminUser{})
	if err != nil {
		t.Fatal(err)
	}

	adapter := postgres_outbound_adapter.NewUserAdapter(pgContainer.DB)

	Convey("Test Postgres User Adapter (Integration)", t, func() {
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.AdminUser{})

		isActive := true
		user := &model.AdminUser{
			FullName:     "Test User",
			Email:        "test@example.com",
			PasswordHash: "hashedpassword",
			Role:         model.AdminRoleCS,
			IsActive:     &isActive,
		}

		Convey("Create", func() {
			Convey("Insert new user", func() {
				err := adapter.Create(user)
				So(err, ShouldBeNil)
				So(user.ID.String(), ShouldNotBeEmpty)

				var count int64
				pgContainer.DB.Model(&model.AdminUser{}).Count(&count)
				So(count, ShouldEqual, 1)
			})
		})

		Convey("FindByEmail", func() {
			adapter.Create(user)

			Convey("Find existing user", func() {
				found, err := adapter.FindByEmail(user.Email)
				So(err, ShouldBeNil)
				So(found.FullName, ShouldEqual, user.FullName)
				So(found.Email, ShouldEqual, user.Email)
			})

			Convey("Find non-existent user", func() {
				_, err := adapter.FindByEmail("nonexistent@example.com")
				So(err, ShouldNotBeNil)
			})
		})

		Convey("FindByID", func() {
			adapter.Create(user)

			Convey("Find existing user by ID", func() {
				found, err := adapter.FindByID(user.ID.String())
				So(err, ShouldBeNil)
				So(found.ID, ShouldEqual, user.ID)
				So(found.FullName, ShouldEqual, user.FullName)
			})

			Convey("Find non-existent user by ID", func() {
				_, err := adapter.FindByID("550e8400-e29b-41d4-a716-446655449999")
				So(err, ShouldNotBeNil)
			})
		})

		Convey("Update", func() {
			adapter.Create(user)

			Convey("Update user data", func() {
				user.FullName = "Updated Name"
				err := adapter.Update(*user)
				So(err, ShouldBeNil)

				found, _ := adapter.FindByID(user.ID.String())
				So(found.FullName, ShouldEqual, "Updated Name")
			})
		})
	})
}
