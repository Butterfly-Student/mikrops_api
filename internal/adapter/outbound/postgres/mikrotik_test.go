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

func TestMikrotikAdapter(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	pgContainer, err := helpers.SetupPostgresContainer(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer pgContainer.Terminate(ctx)

	err = pgContainer.DB.AutoMigrate(&model.MikrotikRouter{})
	if err != nil {
		t.Fatal(err)
	}

	adapter := postgres_outbound_adapter.NewMikrotikAdapter(pgContainer.DB)

	Convey("Test Postgres Mikrotik Adapter (Integration)", t, func() {
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.MikrotikRouter{})

		isActive := true
		useSSL := false
		router := &model.MikrotikRouter{
			Name:     "Test Router",
			Address:  "192.168.1.1:8728",
			Username: "admin",
			Password: "password",
			IsActive: &isActive,
			UseSSL:   &useSSL,
		}

		Convey("Create", func() {
			Convey("Insert new router", func() {
				err := adapter.Create(router)
				So(err, ShouldBeNil)
				So(router.ID.String(), ShouldNotBeEmpty)

				var count int64
				pgContainer.DB.Model(&model.MikrotikRouter{}).Count(&count)
				So(count, ShouldEqual, 1)
			})
		})

		Convey("FindByID", func() {
			adapter.Create(router)

			Convey("Find existing router", func() {
				found, err := adapter.FindByID(router.ID.String())
				So(err, ShouldBeNil)
				So(found.Name, ShouldEqual, router.Name)
				So(found.Address, ShouldEqual, router.Address)
			})

			Convey("Find non-existent router", func() {
				_, err := adapter.FindByID("550e8400-e29b-41d4-a716-446655449999")
				So(err, ShouldNotBeNil)
			})
		})

		Convey("FindAll", func() {
			adapter.Create(router)

			Convey("Find all routers", func() {
				routers, err := adapter.FindAll()
				So(err, ShouldBeNil)
				So(len(routers), ShouldBeGreaterThanOrEqualTo, 1)
			})
		})

		Convey("Update", func() {
			adapter.Create(router)

			Convey("Update router data", func() {
				router.Name = "Updated Router Name"
				err := adapter.Update(router)
				So(err, ShouldBeNil)

				found, _ := adapter.FindByID(router.ID.String())
				So(found.Name, ShouldEqual, "Updated Router Name")
			})
		})

		Convey("Delete", func() {
			adapter.Create(router)

			Convey("Delete router", func() {
				err := adapter.Delete(router.ID.String())
				So(err, ShouldBeNil)

				_, err = adapter.FindByID(router.ID.String())
				So(err, ShouldNotBeNil)
			})
		})
	})
}
