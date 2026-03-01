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

func TestInvoiceAdapter(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	pgContainer, err := helpers.SetupPostgresContainer(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer pgContainer.Terminate(ctx)

	err = pgContainer.DB.AutoMigrate(&model.Invoice{}, &model.InvoiceItem{})
	if err != nil {
		t.Fatal(err)
	}

	adapter := postgres_outbound_adapter.NewInvoiceAdapter(pgContainer.DB)

	Convey("Test Postgres Invoice Adapter (Integration)", t, func() {
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.InvoiceItem{})
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.Invoice{})

		customerID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
		invoice := &model.Invoice{
			InvoiceNumber:      "INV-" + time.Now().Format("20060102150405"),
			CustomerID:         customerID,
			BillingPeriodStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			BillingPeriodEnd:   time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
			DueDate:            time.Date(2024, 2, 5, 0, 0, 0, 0, time.UTC),
			TotalAmount:        100000,
			Status:             model.InvoiceStatusDraft,
		}

		Convey("Create", func() {
			Convey("Insert new invoice", func() {
				err := adapter.Create(ctx, invoice)
				So(err, ShouldBeNil)
				So(invoice.ID.String(), ShouldNotBeEmpty)

				var count int64
				pgContainer.DB.Model(&model.Invoice{}).Count(&count)
				So(count, ShouldEqual, 1)
			})
		})

		Convey("FindByID", func() {
			adapter.Create(ctx, invoice)

			Convey("Find existing invoice", func() {
				found, err := adapter.FindByID(ctx, invoice.ID.String())
				So(err, ShouldBeNil)
				So(found.InvoiceNumber, ShouldEqual, invoice.InvoiceNumber)
				So(found.TotalAmount, ShouldEqual, invoice.TotalAmount)
			})

			Convey("Find non-existent invoice", func() {
				_, err := adapter.FindByID(ctx, "550e8400-e29b-41d4-a716-446655449999")
				So(err, ShouldNotBeNil)
			})
		})

		Convey("FindByNumber", func() {
			adapter.Create(ctx, invoice)

			Convey("Find by invoice number", func() {
				found, err := adapter.FindByNumber(ctx, invoice.InvoiceNumber)
				So(err, ShouldBeNil)
				So(found.ID, ShouldEqual, invoice.ID)
			})
		})

		Convey("FindAll", func() {
			adapter.Create(ctx, invoice)

			Convey("Find all without filter", func() {
				invoices, err := adapter.FindAll(ctx, nil)
				So(err, ShouldBeNil)
				So(len(invoices), ShouldBeGreaterThanOrEqualTo, 1)
			})

			Convey("Find with status filter", func() {
				status := model.InvoiceStatusDraft
				filter := &model.InvoiceFilter{Status: &status}
				invoices, err := adapter.FindAll(ctx, filter)
				So(err, ShouldBeNil)
				So(len(invoices), ShouldBeGreaterThanOrEqualTo, 1)
			})

			Convey("Find with customer filter", func() {
				filter := &model.InvoiceFilter{CustomerID: &customerID}
				invoices, err := adapter.FindAll(ctx, filter)
				So(err, ShouldBeNil)
				So(len(invoices), ShouldBeGreaterThanOrEqualTo, 1)
			})
		})

		Convey("Update", func() {
			adapter.Create(ctx, invoice)

			Convey("Update invoice data", func() {
				invoice.TotalAmount = 150000
				err := adapter.Update(ctx, invoice)
				So(err, ShouldBeNil)

				found, _ := adapter.FindByID(ctx, invoice.ID.String())
				So(found.TotalAmount, ShouldEqual, 150000)
			})
		})

		Convey("Delete", func() {
			adapter.Create(ctx, invoice)

			Convey("Soft delete invoice", func() {
				err := adapter.Delete(ctx, invoice.ID.String())
				So(err, ShouldBeNil)

				_, err = adapter.FindByID(ctx, invoice.ID.String())
				So(err, ShouldNotBeNil)
			})
		})

		Convey("GetLastInvoiceNumber", func() {
			adapter.Create(ctx, invoice)

			Convey("Get last invoice number", func() {
				lastNumber, err := adapter.GetLastInvoiceNumber(ctx, 2024, 1)
				So(err, ShouldBeNil)
				So(lastNumber, ShouldNotBeEmpty)
			})

			Convey("No invoice for month", func() {
				lastNumber, err := adapter.GetLastInvoiceNumber(ctx, 1999, 1)
				So(err, ShouldBeNil)
				So(lastNumber, ShouldBeEmpty)
			})
		})
	})
}
