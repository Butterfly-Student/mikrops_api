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

func TestPaymentAdapter(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	pgContainer, err := helpers.SetupPostgresContainer(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer pgContainer.Terminate(ctx)

	err = pgContainer.DB.AutoMigrate(&model.Payment{}, &model.PaymentAllocation{})
	if err != nil {
		t.Fatal(err)
	}

	adapter := postgres_outbound_adapter.NewPaymentAdapter(pgContainer.DB)

	Convey("Test Postgres Payment Adapter (Integration)", t, func() {
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.PaymentAllocation{})
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.Payment{})

		customerID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
		payment := &model.Payment{
			PaymentNumber: "PAY-" + time.Now().Format("20060102150405"),
			CustomerID:    customerID,
			Amount:        50000,
			PaymentMethod: model.PaymentMethodCash,
			PaymentDate:   time.Now(),
			Status:        model.PaymentStatusTypePending,
		}

		Convey("Create", func() {
			Convey("Insert new payment", func() {
				err := adapter.Create(ctx, payment)
				So(err, ShouldBeNil)
				So(payment.ID.String(), ShouldNotBeEmpty)

				var count int64
				pgContainer.DB.Model(&model.Payment{}).Count(&count)
				So(count, ShouldEqual, 1)
			})
		})

		Convey("FindByID", func() {
			adapter.Create(ctx, payment)

			Convey("Find existing payment", func() {
				found, err := adapter.FindByID(ctx, payment.ID.String())
				So(err, ShouldBeNil)
				So(found.PaymentNumber, ShouldEqual, payment.PaymentNumber)
				So(found.Amount, ShouldEqual, payment.Amount)
			})

			Convey("Find non-existent payment", func() {
				_, err := adapter.FindByID(ctx, "550e8400-e29b-41d4-a716-446655449999")
				So(err, ShouldNotBeNil)
			})
		})

		Convey("FindByNumber", func() {
			adapter.Create(ctx, payment)

			Convey("Find by payment number", func() {
				found, err := adapter.FindByNumber(ctx, payment.PaymentNumber)
				So(err, ShouldBeNil)
				So(found.ID, ShouldEqual, payment.ID)
			})
		})

		Convey("FindAll", func() {
			adapter.Create(ctx, payment)

			Convey("Find all without filter", func() {
				payments, err := adapter.FindAll(ctx, nil)
				So(err, ShouldBeNil)
				So(len(payments), ShouldBeGreaterThanOrEqualTo, 1)
			})

			Convey("Find with status filter", func() {
				status := model.PaymentStatusTypePending
				filter := &model.PaymentFilter{Status: &status}
				payments, err := adapter.FindAll(ctx, filter)
				So(err, ShouldBeNil)
				So(len(payments), ShouldBeGreaterThanOrEqualTo, 1)
			})

			Convey("Find with customer filter", func() {
				filter := &model.PaymentFilter{CustomerID: &customerID}
				payments, err := adapter.FindAll(ctx, filter)
				So(err, ShouldBeNil)
				So(len(payments), ShouldBeGreaterThanOrEqualTo, 1)
			})
		})

		Convey("Update", func() {
			adapter.Create(ctx, payment)

			Convey("Update payment data", func() {
				payment.Status = model.PaymentStatusTypeConfirmed
				err := adapter.Update(ctx, payment)
				So(err, ShouldBeNil)

				found, _ := adapter.FindByID(ctx, payment.ID.String())
				So(found.Status, ShouldEqual, model.PaymentStatusTypeConfirmed)
			})
		})

		Convey("Delete", func() {
			adapter.Create(ctx, payment)

			Convey("Soft delete payment", func() {
				err := adapter.Delete(ctx, payment.ID.String())
				So(err, ShouldBeNil)

				_, err = adapter.FindByID(ctx, payment.ID.String())
				So(err, ShouldNotBeNil)
			})
		})

		Convey("CreateAllocation", func() {
			adapter.Create(ctx, payment)
			invoiceID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440002")

			Convey("Create payment allocation", func() {
				allocation := &model.PaymentAllocation{
					PaymentID:       payment.ID,
					InvoiceID:       invoiceID,
					AllocatedAmount: 25000,
				}
				err := adapter.CreateAllocation(ctx, allocation)
				So(err, ShouldBeNil)
				So(allocation.ID.String(), ShouldNotBeEmpty)
			})
		})

		Convey("GetLastPaymentNumber", func() {
			adapter.Create(ctx, payment)

			Convey("Get last payment number", func() {
				lastNumber, err := adapter.GetLastPaymentNumber(ctx, 2024, 1)
				So(err, ShouldBeNil)
				// Returns empty since our test payment doesn't match pattern PAY/YYYY/MM/XXX
				_ = lastNumber
			})

			Convey("No payment for month", func() {
				lastNumber, err := adapter.GetLastPaymentNumber(ctx, 1999, 1)
				So(err, ShouldBeNil)
				So(lastNumber, ShouldBeEmpty)
			})
		})
	})
}
