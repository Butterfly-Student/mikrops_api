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

func TestSubscriptionIntegration(t *testing.T) {
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
		&model.Customer{},
		&model.BandwidthProfile{},
		&model.MikrotikRouter{},
		&model.Subscription{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate tables: %v", err)
	}

	Convey("Test Subscription Integration with PostgreSQL", t, func() {
		adapter := postgres_outbound_adapter.NewSubscriptionAdapter(pgContainer.DB)
		customerAdapter := postgres_outbound_adapter.NewCustomerAdapter(pgContainer.DB)

		// Cleanup before test
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.Subscription{})
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.Customer{})

		// Create a customer first
		customerInput := model.CustomerInput{
			CustomerCode: "CUST-SUB-" + time.Now().Format("20060102150405"),
			FullName:     "Subscription Test Customer",
			Phone:        "08123456789",
		}
		customer := customerInput.ToModel()
		err := customerAdapter.Create(ctx, &customer)
		So(err, ShouldBeNil)

		Convey("Full CRUD cycle", func() {
			input := model.SubscriptionInput{
				CustomerID:   customer.ID,
				PlanID:       uuid.MustParse("550e8400-e29b-41d4-a716-446655440010"),
				RouterID:     uuid.MustParse("550e8400-e29b-41d4-a716-446655440020"),
				ServiceType:  "pppoe",
				Username:     "testuser",
				Password:     "testpass",
				BillingCycle: strPtr("monthly"),
			}

			Convey("Create creates a new subscription", func() {
				subscription := input.ToModel()
				err := adapter.Create(ctx, &subscription)
				So(err, ShouldBeNil)
				So(subscription.ID.String(), ShouldNotBeEmpty)
				So(subscription.Status, ShouldEqual, model.SubscriptionStatusPending)

				Convey("FindByID retrieves the subscription", func() {
					found, err := adapter.FindByID(ctx, subscription.ID.String())
					So(err, ShouldBeNil)
					So(found.Username, ShouldEqual, "testuser")
					So(found.CustomerID, ShouldEqual, customer.ID)
				})

				Convey("FindByCustomerID retrieves subscriptions", func() {
					subscriptions, err := adapter.FindByCustomerID(ctx, customer.ID.String())
					So(err, ShouldBeNil)
					So(len(subscriptions), ShouldBeGreaterThanOrEqualTo, 1)
				})

				Convey("Update updates the subscription", func() {
					subscription.Username = "updateduser"
					err := adapter.Update(ctx, &subscription)
					So(err, ShouldBeNil)

					found, err := adapter.FindByID(ctx, subscription.ID.String())
					So(err, ShouldBeNil)
					So(found.Username, ShouldEqual, "updateduser")
				})
			})
		})

		Convey("FindByID returns error for non-existent subscription", func() {
			_, err := adapter.FindByID(ctx, "550e8400-e29b-41d4-a716-446655449999")
			So(err, ShouldNotBeNil)
		})
	})
}

func strPtr(s string) *string {
	return &s
}
