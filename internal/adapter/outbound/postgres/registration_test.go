package postgres_outbound_adapter_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
	"gorm.io/gorm"

	postgres_outbound_adapter "go-template/internal/adapter/outbound/postgres"
	"go-template/internal/model"
	"go-template/tests/helpers"
)

func TestRegistrationAdapter(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	pgContainer, err := helpers.SetupPostgresContainer(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer pgContainer.Terminate(ctx)

	err = pgContainer.DB.AutoMigrate(&model.CustomerRegistration{}, &model.Customer{})
	if err != nil {
		t.Fatal(err)
	}

	adapter := postgres_outbound_adapter.NewRegistrationAdapter(pgContainer.DB)

	Convey("Test Postgres Registration Adapter (Integration)", t, func() {
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.CustomerRegistration{})
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.Customer{})

		bandwidthProfileID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
		email := "test@example.com"
		address := "Test Address"

		registration := &model.CustomerRegistration{
			FullName:           "Test Registration",
			Phone:              "08123456789",
			Email:              &email,
			Address:            &address,
			BandwidthProfileID: &bandwidthProfileID,
			Status:             model.RegistrationStatusPending,
		}

		Convey("Create", func() {
			Convey("Insert new registration", func() {
				err := adapter.Create(ctx, registration)
				So(err, ShouldBeNil)
				So(registration.ID.String(), ShouldNotBeEmpty)

				var count int64
				pgContainer.DB.Model(&model.CustomerRegistration{}).Count(&count)
				So(count, ShouldEqual, 1)
			})
		})

		Convey("FindByID", func() {
			adapter.Create(ctx, registration)

			Convey("Find existing registration", func() {
				found, err := adapter.FindByID(ctx, registration.ID.String())
				So(err, ShouldBeNil)
				So(found.FullName, ShouldEqual, registration.FullName)
				So(found.Status, ShouldEqual, registration.Status)
			})

			Convey("Find non-existent registration", func() {
				_, err := adapter.FindByID(ctx, "550e8400-e29b-41d4-a716-446655449999")
				So(err, ShouldNotBeNil)
			})
		})

		Convey("FindAll", func() {
			adapter.Create(ctx, registration)

			Convey("Find all without filter", func() {
				regs, err := adapter.FindAll(ctx, nil)
				So(err, ShouldBeNil)
				So(len(regs), ShouldBeGreaterThanOrEqualTo, 1)
			})

			Convey("Find with status filter", func() {
				status := model.RegistrationStatusPending
				filter := &model.RegistrationFilter{Status: &status}
				regs, err := adapter.FindAll(ctx, filter)
				So(err, ShouldBeNil)
				So(len(regs), ShouldBeGreaterThanOrEqualTo, 1)
			})

			Convey("Find with search filter", func() {
				search := "Test"
				filter := &model.RegistrationFilter{Search: &search}
				regs, err := adapter.FindAll(ctx, filter)
				So(err, ShouldBeNil)
				So(len(regs), ShouldBeGreaterThanOrEqualTo, 1)
			})
		})

		Convey("SetApproved", func() {
			adapter.Create(ctx, registration)

			Convey("Approve registration", func() {
				customerID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440002")
				approverID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440003")
				err := adapter.SetApproved(ctx, registration.ID.String(), approverID.String(), customerID.String())
				So(err, ShouldBeNil)

				found, _ := adapter.FindByID(ctx, registration.ID.String())
				So(found.Status, ShouldEqual, model.RegistrationStatusApproved)
			})
		})

		Convey("SetRejected", func() {
			adapter.Create(ctx, registration)

			Convey("Reject registration", func() {
				approverID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440003")
				err := adapter.SetRejected(ctx, registration.ID.String(), approverID.String(), "Invalid documents")
				So(err, ShouldBeNil)

				found, _ := adapter.FindByID(ctx, registration.ID.String())
				So(found.Status, ShouldEqual, model.RegistrationStatusRejected)
			})
		})
	})
}
