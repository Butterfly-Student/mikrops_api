package postgres_outbound_adapter_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
	"gorm.io/gorm"

	postgres_outbound_adapter "go-template/internal/adapter/outbound/postgres"
	"go-template/internal/model"
	"go-template/tests/helpers"
)

func TestSubscriptionAdapter(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	pgContainer, err := helpers.SetupPostgresContainer(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer pgContainer.Terminate(ctx)

	err = pgContainer.DB.AutoMigrate(&model.Subscription{})
	if err != nil {
		t.Fatal(err)
	}

	adapter := postgres_outbound_adapter.NewSubscriptionAdapter(pgContainer.DB)

	Convey("Test Postgres Subscription Adapter (Integration)", t, func() {
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.Subscription{})

		customerID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
		planID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440002")
		routerID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440003")
		billingDay := 1
		activatedAt := time.Now()
		expiredAt := time.Now().Add(30 * 24 * time.Hour)

		subscription := &model.Subscription{
			CustomerID:    customerID,
			PlanID:        planID,
			RouterID:      routerID,
			Username:      "testuser1",
			Password:      "testpass123",
			ServiceType:   model.ServiceTypePPPoE,
			Status:        model.SubscriptionStatusActive,
			ActivatedAt:   &activatedAt,
			ExpiredAt:     &expiredAt,
			BillingDay:    &billingDay,
			BillingCycle:  model.BillingCycleMonthly,
		}

		Convey("Create", func() {
			Convey("Insert new subscription", func() {
				err := adapter.Create(ctx, subscription)
				So(err, ShouldBeNil)
				So(subscription.ID.String(), ShouldNotBeEmpty)

				var count int64
				pgContainer.DB.Model(&model.Subscription{}).Count(&count)
				So(count, ShouldEqual, 1)
			})
		})

		Convey("FindByID", func() {
			adapter.Create(ctx, subscription)

			Convey("Find existing subscription", func() {
				found, err := adapter.FindByID(ctx, subscription.ID.String())
				So(err, ShouldBeNil)
				So(found.Username, ShouldEqual, subscription.Username)
				So(found.Status, ShouldEqual, subscription.Status)
			})

			Convey("Find non-existent subscription", func() {
				_, err := adapter.FindByID(ctx, "550e8400-e29b-41d4-a716-446655449999")
				So(err, ShouldNotBeNil)
			})
		})

		Convey("FindByCustomerID", func() {
			adapter.Create(ctx, subscription)

			Convey("Find by customer id", func() {
				subs, err := adapter.FindByCustomerID(ctx, customerID.String())
				So(err, ShouldBeNil)
				So(len(subs), ShouldBeGreaterThanOrEqualTo, 1)
			})
		})

		Convey("FindActiveByCustomerID", func() {
			adapter.Create(ctx, subscription)

			Convey("Find active by customer id", func() {
				subs, err := adapter.FindActiveByCustomerID(ctx, customerID.String())
				So(err, ShouldBeNil)
				So(len(subs), ShouldBeGreaterThanOrEqualTo, 1)
			})
		})

		Convey("Update", func() {
			adapter.Create(ctx, subscription)

			Convey("Update subscription data", func() {
				subscription.Status = model.SubscriptionStatusSuspended
				err := adapter.Update(ctx, subscription)
				So(err, ShouldBeNil)

				found, _ := adapter.FindByID(ctx, subscription.ID.String())
				So(found.Status, ShouldEqual, model.SubscriptionStatusSuspended)
			})
		})

		Convey("ListUsernames", func() {
			adapter.Create(ctx, subscription)

			Convey("List all usernames", func() {
				usernames, err := adapter.ListUsernames(ctx)
				So(err, ShouldBeNil)
				So(len(usernames), ShouldBeGreaterThanOrEqualTo, 1)
				So(usernames, ShouldContain, "testuser1")
			})
		})
	})
}
