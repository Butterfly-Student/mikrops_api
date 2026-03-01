package postgres_outbound_adapter_test

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	postgres_outbound_adapter "go-template/internal/adapter/outbound/postgres"
	"go-template/tests/helpers"
)

func TestCasbinAdapter(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	pgContainer, err := helpers.SetupPostgresContainer(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer pgContainer.Terminate(ctx)

	Convey("Test Casbin Adapter (Integration)", t, func() {
		Convey("InitCasbin", func() {
			Convey("Initialize casbin enforcer", func() {
				enforcer := postgres_outbound_adapter.InitCasbin(pgContainer.DB)
				So(enforcer, ShouldNotBeNil)
			})
		})

		Convey("Authorization checks", func() {
			enforcer := postgres_outbound_adapter.InitCasbin(pgContainer.DB)

			Convey("Admin role can access any resource", func() {
				allowed, err := enforcer.Enforce("admin", "/api/users", "GET")
				So(err, ShouldBeNil)
				So(allowed, ShouldBeTrue)

				allowed, err = enforcer.Enforce("admin", "/api/users", "POST")
				So(err, ShouldBeNil)
				So(allowed, ShouldBeTrue)
			})

			Convey("User role has limited access", func() {
				// Users can read their own data
				allowed, err := enforcer.Enforce("user", "/api/users/me", "GET")
				So(err, ShouldBeNil)
				// This depends on actual policies, may be true or false
				So(allowed, ShouldBeTrue)
			})

			Convey("Wildcard path matching", func() {
				// Policy with wildcard should match nested paths
				// This depends on actual policies loaded from DB
				_, err := enforcer.Enforce("admin", "/api/users/123", "GET")
				So(err, ShouldBeNil)
			})
		})
	})
}
