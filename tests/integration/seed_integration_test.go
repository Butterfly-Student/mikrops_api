//go:build integration
// +build integration

package integration_test

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"go-template/internal/model"
	_ "go-template/internal/seeds"
	"go-template/tests/helpers"
)

func TestSeedDataIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	pgContainer, err := helpers.SetupPostgresContainer(ctx)
	if err != nil {
		t.Fatalf("Failed to setup postgres container: %v", err)
	}
	defer pgContainer.Terminate(ctx)

	Convey("Test Seed Data Integration", t, func() {
		Convey("Test users are seeded", func() {
			var users []model.User
			err := pgContainer.DB.Find(&users).Error
			So(err, ShouldBeNil)
			So(len(users), ShouldBeGreaterThanOrEqualTo, 4)

			var testAdmin model.User
			err = pgContainer.DB.Where("email = ?", "test-admin@example.com").First(&testAdmin).Error
			So(err, ShouldBeNil)
			So(testAdmin.Name, ShouldEqual, "Test Admin")
			So(testAdmin.Role, ShouldEqual, "admin")
			So(testAdmin.Status, ShouldEqual, "active")
		})

		Convey("Test clients are seeded", func() {
			var clients []model.Client
			err := pgContainer.DB.Find(&clients).Error
			So(err, ShouldBeNil)
			So(len(clients), ShouldBeGreaterThanOrEqualTo, 3)

			var testClient model.Client
			err = pgContainer.DB.Where("name = ?", "test-client-1").First(&testClient).Error
			So(err, ShouldBeNil)
			So(testClient.Name, ShouldEqual, "test-client-1")
		})

		Convey("Test MikroTik routers are seeded", func() {
			var routers []model.MikrotikRouter
			err := pgContainer.DB.Find(&routers).Error
			So(err, ShouldBeNil)
			So(len(routers), ShouldEqual, 3)

			var testRouter model.MikrotikRouter
			err = pgContainer.DB.Where("name = ?", "Test Router 1").First(&testRouter).Error
			So(err, ShouldBeNil)
			So(testRouter.Name, ShouldEqual, "Test Router 1")
			So(testRouter.Address, ShouldEqual, "192.168.1.1:8728")
		})
	})
}
