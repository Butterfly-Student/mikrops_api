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

func TestCustomerIntegration(t *testing.T) {
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

	// Use GORM AutoMigrate
	err = pgContainer.DB.AutoMigrate(&model.Customer{})
	if err != nil {
		t.Fatalf("Failed to migrate table: %v", err)
	}

	Convey("Test Customer Integration with PostgreSQL", t, func() {
		adapter := postgres_outbound_adapter.NewCustomerAdapter(pgContainer.DB)

		// Cleanup before test
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.Customer{})

		Convey("Full CRUD cycle", func() {
			customerCode := "CUST-" + time.Now().Format("20060102150405")
			input := model.CustomerInput{
				CustomerCode: customerCode,
				FullName:     "Integration Customer",
				Email:        strPtr("customer@example.com"),
				Phone:        "08123456789",
				Address:      strPtr("Test Address"),
			}

			Convey("Create creates a new customer", func() {
				customer := input.ToModel()
				err := adapter.Create(ctx, &customer)
				So(err, ShouldBeNil)
				// UUID should be set
				So(customer.ID.String(), ShouldNotBeEmpty)
				// Default status should be pending
				So(customer.Status, ShouldEqual, model.CustomerStatusPending)

				Convey("FindByID retrieves the customer", func() {
					found, err := adapter.FindByID(ctx, customer.ID.String())
					So(err, ShouldBeNil)
					So(found.CustomerCode, ShouldEqual, customerCode)
					So(found.FullName, ShouldEqual, "Integration Customer")
				})

				Convey("FindByCode retrieves the customer", func() {
					found, err := adapter.FindByCode(ctx, customerCode)
					So(err, ShouldBeNil)
					So(found.ID, ShouldEqual, customer.ID)
				})

				Convey("Update updates the customer", func() {
					customer.FullName = "Updated Customer Name"
					err := adapter.Update(ctx, &customer)
					So(err, ShouldBeNil)

					found, err := adapter.FindByID(ctx, customer.ID.String())
					So(err, ShouldBeNil)
					So(found.FullName, ShouldEqual, "Updated Customer Name")
				})

				Convey("UpdateStatus updates the customer status", func() {
					err := adapter.UpdateStatus(ctx, customer.ID.String(), model.CustomerStatusActive)
					So(err, ShouldBeNil)

					found, err := adapter.FindByID(ctx, customer.ID.String())
					So(err, ShouldBeNil)
					So(found.Status, ShouldEqual, model.CustomerStatusActive)
				})

				Convey("Delete soft-deletes the customer", func() {
					err := adapter.Delete(ctx, customer.ID.String())
					So(err, ShouldBeNil)

					// Should not be found after delete
					_, err = adapter.FindByID(ctx, customer.ID.String())
					So(err, ShouldNotBeNil)
				})
			})
		})

		Convey("List customers with filter", func() {
			// Create multiple customers
			for i := 0; i < 3; i++ {
				input := model.CustomerInput{
					CustomerCode: "CUST-LIST-" + time.Now().Format("20060102150405") + "-" + string(rune('0'+i)),
					FullName:     "List Customer " + string(rune('0'+i)),
					Phone:        "0812345678" + string(rune('0'+i)),
				}
				customer := input.ToModel()
				adapter.Create(ctx, &customer)
			}

			Convey("FindAll returns customers", func() {
				customers, err := adapter.FindAll(ctx, nil)
				So(err, ShouldBeNil)
				So(len(customers), ShouldBeGreaterThanOrEqualTo, 3)
			})

			Convey("FindAll with status filter", func() {
				status := model.CustomerStatusPending
				filter := &model.CustomerFilter{
					Status: &status,
				}
				customers, err := adapter.FindAll(ctx, filter)
				So(err, ShouldBeNil)
				So(len(customers), ShouldBeGreaterThanOrEqualTo, 3)
			})
		})

		Convey("FindByCode returns error for non-existent customer", func() {
			_, err := adapter.FindByCode(ctx, "NONEXISTENT")
			So(err, ShouldNotBeNil)
		})

		Convey("FindByID returns error for non-existent customer", func() {
			_, err := adapter.FindByID(ctx, "550e8400-e29b-41d4-a716-446655449999")
			So(err, ShouldNotBeNil)
		})
	})
}

func strPtr(s string) *string {
	return &s
}
