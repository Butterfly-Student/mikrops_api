package postgres_outbound_adapter_test

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"gorm.io/gorm"

	postgres_outbound_adapter "go-template/internal/adapter/outbound/postgres"
	"go-template/internal/model"
	"go-template/tests/helpers"
)

func TestCustomerAdapter(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	pgContainer, err := helpers.SetupPostgresContainer(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer pgContainer.Terminate(ctx)

	err = pgContainer.DB.AutoMigrate(&model.Customer{})
	if err != nil {
		t.Fatal(err)
	}

	adapter := postgres_outbound_adapter.NewCustomerAdapter(pgContainer.DB)

	Convey("Test Postgres Customer Adapter (Integration)", t, func() {
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.Customer{})

		input := model.CustomerInput{
			CustomerCode: "CUST-" + time.Now().Format("20060102150405"),
			FullName:     "Test Customer",
			Email:        strPtr("customer@example.com"),
			Phone:        "08123456789",
			Address:      strPtr("Test Address"),
		}
		customer := input.ToModel()

		Convey("Create", func() {
			Convey("Insert new customer", func() {
				err := adapter.Create(ctx, customer)
				So(err, ShouldBeNil)
				So(customer.ID.String(), ShouldNotBeEmpty)
				So(customer.Status, ShouldEqual, model.CustomerStatusPending)

				var count int64
				pgContainer.DB.Model(&model.Customer{}).Count(&count)
				So(count, ShouldEqual, 1)
			})
		})

		Convey("FindByID", func() {
			adapter.Create(ctx, customer)

			Convey("Find existing customer", func() {
				found, err := adapter.FindByID(ctx, customer.ID.String())
				So(err, ShouldBeNil)
				So(found.CustomerCode, ShouldEqual, customer.CustomerCode)
				So(found.FullName, ShouldEqual, customer.FullName)
			})

			Convey("Find non-existent customer", func() {
				_, err := adapter.FindByID(ctx, "550e8400-e29b-41d4-a716-446655449999")
				So(err, ShouldNotBeNil)
			})
		})

		Convey("FindByCode", func() {
			adapter.Create(ctx, customer)

			Convey("Find by customer code", func() {
				found, err := adapter.FindByCode(ctx, customer.CustomerCode)
				So(err, ShouldBeNil)
				So(found.ID, ShouldEqual, customer.ID)
			})
		})

		Convey("FindByPortalIdentifier", func() {
			adapter.Create(ctx, customer)

			Convey("Find by phone", func() {
				found, err := adapter.FindByPortalIdentifier(ctx, customer.Phone)
				So(err, ShouldBeNil)
				So(found.ID, ShouldEqual, customer.ID)
			})

			Convey("Find by customer code", func() {
				found, err := adapter.FindByPortalIdentifier(ctx, customer.CustomerCode)
				So(err, ShouldBeNil)
				So(found.ID, ShouldEqual, customer.ID)
			})
		})

		Convey("FindAll", func() {
			adapter.Create(ctx, customer)

			Convey("Find all without filter", func() {
				customers, err := adapter.FindAll(ctx, nil)
				So(err, ShouldBeNil)
				So(len(customers), ShouldBeGreaterThanOrEqualTo, 1)
			})

			Convey("Find with status filter", func() {
				status := model.CustomerStatusPending
				filter := &model.CustomerFilter{Status: &status}
				customers, err := adapter.FindAll(ctx, filter)
				So(err, ShouldBeNil)
				So(len(customers), ShouldBeGreaterThanOrEqualTo, 1)
			})

			Convey("Find with search filter", func() {
				search := customer.FullName
				filter := &model.CustomerFilter{Search: &search}
				customers, err := adapter.FindAll(ctx, filter)
				So(err, ShouldBeNil)
				So(len(customers), ShouldBeGreaterThanOrEqualTo, 1)
			})
		})

		Convey("Update", func() {
			adapter.Create(ctx, customer)

			Convey("Update customer data", func() {
				customer.FullName = "Updated Customer Name"
				err := adapter.Update(ctx, customer)
				So(err, ShouldBeNil)

				found, _ := adapter.FindByID(ctx, customer.ID.String())
				So(found.FullName, ShouldEqual, "Updated Customer Name")
			})
		})

		Convey("UpdateStatus", func() {
			adapter.Create(ctx, customer)

			Convey("Update to active", func() {
				err := adapter.UpdateStatus(ctx, customer.ID.String(), model.CustomerStatusActive)
				So(err, ShouldBeNil)

				found, _ := adapter.FindByID(ctx, customer.ID.String())
				So(found.Status, ShouldEqual, model.CustomerStatusActive)
			})
		})

		Convey("Delete", func() {
			adapter.Create(ctx, customer)

			Convey("Soft delete customer", func() {
				err := adapter.Delete(ctx, customer.ID.String())
				So(err, ShouldBeNil)

				_, err = adapter.FindByID(ctx, customer.ID.String())
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func strPtr(s string) *string {
	return &s
}
