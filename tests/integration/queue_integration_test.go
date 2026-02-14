//go:build integration
// +build integration

package integration_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	. "github.com/smartystreets/goconvey/convey"
	"gorm.io/gorm"

	postgres_outbound_adapter "go-template/internal/adapter/outbound/postgres"
	"go-template/internal/model"
	"go-template/tests/helpers"
)

func TestQueueIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	pgContainer, err := helpers.SetupPostgresContainer(ctx)
	if err != nil {
		t.Fatalf("Failed to setup postgres container: %v", err)
	}
	defer pgContainer.Terminate(ctx)

	err = pgContainer.DB.AutoMigrate(
		&model.PppoeQueue{},
		&model.MikrotikRouter{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate queue tables: %v", err)
	}

	mikrotikAdapter := postgres_outbound_adapter.NewMikrotikAdapter(pgContainer.DB)

	Convey("Test Queue Integration with PostgreSQL", t, func() {
		routerID := uuid.New()
		router := &model.MikrotikRouter{
			ID:       routerID,
			Name:     "Queue Integration Router",
			Address:  "192.168.88.10:8728",
			Username: "admin",
			Password: "admin",
			IsActive: &[]bool{true}[0],
		}

		err := mikrotikAdapter.Create(router)
		So(err, ShouldBeNil)

		defer func() {
			pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.PppoeQueue{})
			pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.MikrotikRouter{})
		}()

		Convey("Router operations", func() {
			Convey("FindByID retrieves router", func() {
				found, err := mikrotikAdapter.FindByID(routerID.String())
				So(err, ShouldBeNil)
				So(found.ID, ShouldEqual, routerID)
				So(found.Name, ShouldEqual, "Queue Integration Router")
			})

			Convey("Update router fields", func() {
				router.Name = "Updated Queue Router"
				err := mikrotikAdapter.Update(router)
				So(err, ShouldBeNil)

				found, err := mikrotikAdapter.FindByID(routerID.String())
				So(err, ShouldBeNil)
				So(found.Name, ShouldEqual, "Updated Queue Router")
			})

			Convey("FindAll returns all routers", func() {
				router2 := &model.MikrotikRouter{
					ID:       uuid.New(),
					Name:     "Second Queue Router",
					Address:  "192.168.88.11:8728",
					Username: "admin",
					Password: "admin",
					IsActive: &[]bool{true}[0],
				}
				err := mikrotikAdapter.Create(router2)
				So(err, ShouldBeNil)

				routers, err := mikrotikAdapter.FindAll()
				So(err, ShouldBeNil)
				So(len(routers), ShouldBeGreaterThanOrEqualTo, 2)

				pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(router2)
			})
		})
	})
}
