//go:build integration
// +build integration

package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	. "github.com/smartystreets/goconvey/convey"
	"gorm.io/gorm"

	postgres_outbound_adapter "go-template/internal/adapter/outbound/postgres"
	"go-template/internal/domain"
	"go-template/internal/model"
	"go-template/tests/helpers"
)

func TestBillingIntegration(t *testing.T) {
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
		&model.Invoice{},
		&model.InvoiceItem{},
		&model.Customer{},
		&model.BandwidthProfile{},
		&model.SystemSetting{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate billing tables: %v", err)
	}

	dbAdapter := postgres_outbound_adapter.NewAdapter(pgContainer.DB)
	systemSettingAdapter := postgres_outbound_adapter.NewSystemSettingAdapter(pgContainer.DB)
	billingDomain := domain.NewBillingDomain(dbAdapter, systemSettingAdapter)

	Convey("Test Billing Integration with PostgreSQL", t, func() {
		// Cleanup before test
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.InvoiceItem{})
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.Invoice{})
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.Customer{})
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.BandwidthProfile{})
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.SystemSetting{})

		Convey("Setup test data", func() {
			// Create bandwidth profile
			profile := &model.BandwidthProfile{
				Name:                "Test Profile",
				Category:            "pppoe",
				PriceMonthly:        100000,
				TaxRate:             0.11,
				DownloadSpeed:       "10Mbps",
				UploadSpeed:         "5Mbps",
				IsActive:            true,
			}
			err := dbAdapter.BandwidthProfile().Create(profile)
			So(err, ShouldBeNil)

			// Create customer
			customerCode := "BILL-" + time.Now().Format("20060102150405")
			expiryDate := time.Now().AddDate(0, 1, 0)
			customer := &model.Customer{
				CustomerCode: customerCode,
				FullName:     "Billing Test Customer",
				Email:        func() *string { s := "billing@test.com"; return &s }(),
				Phone:        "08123456789",
				Address:      func() *string { s := "Test Address"; return &s }(),
				Status:       "active",
				ExpiryDate:   &expiryDate,
				ProfileID:    &profile.ID,
			}
			err = dbAdapter.Customer().Create(customer)
			So(err, ShouldBeNil)

			Convey("CreateInvoice creates invoice with items", func() {
				billingStart := time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC)
				billingEnd := time.Date(2026, time.February, 28, 23, 59, 59, 0, time.UTC)
				issueDate := time.Now()
				dueDate := issueDate.AddDate(0, 0, 7)
				billingMonth := 2
				billingYear := 2026

				itemDesc := "Test Subscription Service"
				itemType := "subscription"

				input := model.InvoiceInput{
					CustomerID:         customer.ID,
					BillingPeriodStart: billingStart,
					BillingPeriodEnd:   billingEnd,
					BillingMonth:       &billingMonth,
					BillingYear:        &billingYear,
					IssueDate:          &issueDate,
					DueDate:            &dueDate,
					Subtotal:           func() *float64 { f := 100000.0; return &f }(),
					TaxAmount:          func() *float64 { f := 11000.0; return &f }(),
					TotalAmount:        func() *float64 { f := 111000.0; return &f }(),
					Status:             func() *string { s := "draft"; return &s }(),
					Items: []model.InvoiceItemInput{
						{
							ItemType:    itemType,
							Description: itemDesc,
							ProfileID:   &profile.ID,
							Quantity:    func() *int { i := 1; return &i }(),
							UnitPrice:   100000.0,
						},
					},
				}

				invoice, err := billingDomain.CreateInvoice(ctx, input)
				So(err, ShouldBeNil)
				So(invoice, ShouldNotBeNil)
				So(invoice.InvoiceNumber, ShouldNotBeEmpty)
				So(invoice.CustomerID, ShouldEqual, customer.ID)
				So(invoice.Subtotal, ShouldEqual, 100000.0)
				So(invoice.TaxAmount, ShouldEqual, 11000.0)
				So(invoice.TotalAmount, ShouldEqual, 111000.0)
				So(invoice.Status, ShouldEqual, "draft")

				Convey("GetInvoice retrieves created invoice", func() {
					found, err := billingDomain.GetInvoice(ctx, invoice.ID.String())
					So(err, ShouldBeNil)
					So(found.ID, ShouldEqual, invoice.ID)
					So(found.InvoiceNumber, ShouldEqual, invoice.InvoiceNumber)
				})

				Convey("ListInvoices returns invoices", func() {
					filter := model.InvoiceFilter{
						CustomerIDs: []uuid.UUID{customer.ID},
					}
					invoices, err := billingDomain.ListInvoices(ctx, filter)
					So(err, ShouldBeNil)
					So(len(invoices), ShouldBeGreaterThanOrEqualTo, 1)
				})

				Convey("UpdateInvoice updates invoice fields", func() {
					newStatus := "sent"
					input := model.InvoiceInput{
						Status: &newStatus,
					}
					updated, err := billingDomain.UpdateInvoice(ctx, invoice.ID.String(), input)
					So(err, ShouldBeNil)
					So(updated.Status, ShouldEqual, "sent")
				})

				Convey("ApplyPayment updates paid amount", func() {
					err := billingDomain.ApplyPayment(ctx, invoice.ID.String(), 50000.0)
					So(err, ShouldBeNil)

					updated, err := billingDomain.GetInvoice(ctx, invoice.ID.String())
					So(err, ShouldBeNil)
					So(updated.PaidAmount, ShouldEqual, 50000.0)
					So(updated.PaymentStatus, ShouldEqual, "partial")
				})

				Convey("CreateInvoiceItem creates additional item", func() {
					itemInput := model.InvoiceItemInput{
						ItemType:    "installation",
						Description: "Installation Fee",
						ProfileID:   &profile.ID,
						Quantity:    func() *int { i := 1; return &i }(),
						UnitPrice:   50000.0,
					}
					item, err := billingDomain.CreateInvoiceItem(ctx, itemInput)
					So(err, ShouldBeNil)
					So(item.Description, ShouldEqual, "Installation Fee")
					So(item.Subtotal, ShouldEqual, 50000.0)
				})

				Convey("ListInvoiceItems retrieves items", func() {
					items, err := billingDomain.ListInvoiceItems(ctx, invoice.ID.String())
					So(err, ShouldBeNil)
					So(len(items), ShouldBeGreaterThanOrEqualTo, 1)
				})
			})

			Convey("CreateInvoice without items returns error", func() {
				billingStart := time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC)
				billingEnd := time.Date(2026, time.February, 28, 23, 59, 59, 0, time.UTC)
				issueDate := time.Now()
				dueDate := issueDate.AddDate(0, 0, 7)

				input := model.InvoiceInput{
					CustomerID:         customer.ID,
					BillingPeriodStart: billingStart,
					BillingPeriodEnd:   billingEnd,
					IssueDate:          &issueDate,
					DueDate:            &dueDate,
					Items:              []model.InvoiceItemInput{},
				}

				_, err := billingDomain.CreateInvoice(ctx, input)
				So(err, ShouldNotBeNil)
			})

			Convey("CreateInvoice with non-existent customer returns error", func() {
				billingStart := time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC)
				billingEnd := time.Date(2026, time.February, 28, 23, 59, 59, 0, time.UTC)
				issueDate := time.Now()
				dueDate := issueDate.AddDate(0, 0, 7)

				input := model.InvoiceInput{
					CustomerID:         uuid.New(),
					BillingPeriodStart: billingStart,
					BillingPeriodEnd:   billingEnd,
					IssueDate:          &issueDate,
					DueDate:            &dueDate,
					Items: []model.InvoiceItemInput{
						{
							ItemType:    "subscription",
							Description: "Test",
							ProfileID:   &profile.ID,
							Quantity:    func() *int { i := 1; return &i }(),
							UnitPrice:   100000.0,
						},
					},
				}

				_, err := billingDomain.CreateInvoice(ctx, input)
				So(err, ShouldNotBeNil)
			})

			Convey("CalculateLateFee calculates correctly", func() {
				// Create system setting for late fee
				enabledSetting := &model.SystemSetting{
					Key:   "invoice.late_fee_enabled",
					Value: func() *string { s := "true"; return &s }(),
				}
				err := systemSettingAdapter.Create(enabledSetting)
				So(err, ShouldBeNil)

				amountSetting := &model.SystemSetting{
					Key:   "invoice.late_fee_amount",
					Value: func() *string { s := "5000"; return &s }(),
				}
				err = systemSettingAdapter.Create(amountSetting)
				So(err, ShouldBeNil)

				// Create overdue invoice
				billingStart := time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC)
				billingEnd := time.Date(2026, time.February, 28, 23, 59, 59, 0, time.UTC)
				issueDate := time.Now().AddDate(0, -1, 0)
				dueDate := issueDate.AddDate(0, 0, 7)

				input := model.InvoiceInput{
					CustomerID:         customer.ID,
					BillingPeriodStart: billingStart,
					BillingPeriodEnd:   billingEnd,
					IssueDate:          &issueDate,
					DueDate:            &dueDate,
					Items: []model.InvoiceItemInput{
						{
							ItemType:    "subscription",
							Description: "Test Service",
							ProfileID:   &profile.ID,
							Quantity:    func() *int { i := 1; return &i }(),
							UnitPrice:   100000.0,
						},
					},
				}

				invoice, err := billingDomain.CreateInvoice(ctx, input)
				So(err, ShouldBeNil)

				lateFee := billingDomain.CalculateLateFee(ctx, invoice)
				So(lateFee, ShouldBeGreaterThan, 0)
			})
		})

		Convey("Error handling", func() {
			Convey("GetInvoice with invalid ID returns error", func() {
				_, err := billingDomain.GetInvoice(ctx, uuid.New().String())
				So(err, ShouldNotBeNil)
			})

			Convey("UpdateInvoice with invalid ID returns error", func() {
				input := model.InvoiceInput{
					Status: func() *string { s := "sent"; return &s }(),
				}
				_, err := billingDomain.UpdateInvoice(ctx, uuid.New().String(), input)
				So(err, ShouldNotBeNil)
			})

			Convey("DeleteInvoice with non-empty items returns error", func() {
				// First create an invoice with items
				profile := &model.BandwidthProfile{
					Name:         "Delete Test Profile",
					Category:     "pppoe",
					PriceMonthly: 100000,
					TaxRate:      0.11,
					IsActive:     true,
				}
				err := dbAdapter.BandwidthProfile().Create(profile)
				So(err, ShouldBeNil)

				customerCode := "DEL-" + time.Now().Format("20060102150405")
				expiryDate := time.Now().AddDate(0, 1, 0)
				customer := &model.Customer{
					CustomerCode: customerCode,
					FullName:     "Delete Test Customer",
					Email:        func() *string { s := "delete@test.com"; return &s }(),
					Phone:        "08123456789",
					Status:       "active",
					ExpiryDate:   &expiryDate,
					ProfileID:    &profile.ID,
				}
				err = dbAdapter.Customer().Create(customer)
				So(err, ShouldBeNil)

				billingStart := time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC)
				billingEnd := time.Date(2026, time.February, 28, 23, 59, 59, 0, time.UTC)
				issueDate := time.Now()
				dueDate := issueDate.AddDate(0, 0, 7)

				input := model.InvoiceInput{
					CustomerID:         customer.ID,
					BillingPeriodStart: billingStart,
					BillingPeriodEnd:   billingEnd,
					IssueDate:          &issueDate,
					DueDate:            &dueDate,
					Items: []model.InvoiceItemInput{
						{
							ItemType:    "subscription",
							Description: "Test",
							ProfileID:   &profile.ID,
							Quantity:    func() *int { i := 1; return &i }(),
							UnitPrice:   100000.0,
						},
					},
				}

				invoice, err := billingDomain.CreateInvoice(ctx, input)
				So(err, ShouldBeNil)

				// Try to delete invoice with items
				err = billingDomain.DeleteInvoice(ctx, invoice.ID.String())
				So(err, ShouldNotBeNil)
			})
		})
	})
}
