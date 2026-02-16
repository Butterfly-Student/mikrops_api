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
	"go-template/internal/domain/payment"
	"go-template/internal/model"
	"go-template/tests/helpers"
	"go-template/utils/xendit"
)

func TestPaymentIntegration(t *testing.T) {
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
		&model.Payment{},
		&model.PaymentAllocation{},
		&model.Invoice{},
		&model.InvoiceItem{},
		&model.Customer{},
		&model.BandwidthProfile{},
		&model.SystemSetting{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate payment tables: %v", err)
	}

	dbAdapter := postgres_outbound_adapter.NewAdapter(pgContainer.DB)
	paymentDomain := payment.NewPaymentDomain(dbAdapter, nil)

	Convey("Test Payment Integration with PostgreSQL", t, func() {
		// Cleanup before test
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.PaymentAllocation{})
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.Payment{})
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.InvoiceItem{})
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.Invoice{})
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.Customer{})
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.BandwidthProfile{})

		Convey("Setup test data", func() {
			// Create bandwidth profile
			isActive := true
			profile := &model.BandwidthProfile{
				Name:         "Test Payment Profile",
				Category:     "pppoe",
				PriceMonthly: 100000,
				TaxRate:      0.11,
				IsActive:     &isActive,
			}
			err := dbAdapter.BandwidthProfile().Create(profile)
			So(err, ShouldBeNil)

			// Create customer
			customerCode := "PAY-" + time.Now().Format("20060102150405")
			expiryDate := time.Now().AddDate(0, 1, 0)
			customerEmail := "payment@test.com"
			customer := &model.Customer{
				CustomerCode: customerCode,
				FullName:     "Payment Test Customer",
				Email:        &customerEmail,
				Phone:        "08123456789",
				Status:       "active",
				ExpiryDate:   &expiryDate,
				ProfileID:    &profile.ID,
			}
			err = dbAdapter.Customer().Create(customer)
			So(err, ShouldBeNil)

			Convey("CreatePayment creates payment successfully", func() {
				paymentDate := time.Now()
				amount := 100000.0
				paymentMethod := "bank_transfer"
				bankName := "BCA"
				bankAccountNumber := "1234567890"
				bankAccountName := "Test Account"

				input := model.PaymentInput{
					CustomerID:        customer.ID,
					Amount:            &amount,
					PaymentMethod:     &paymentMethod,
					PaymentDate:       &paymentDate,
					BankName:          &bankName,
					BankAccountNumber: &bankAccountNumber,
					BankAccountName:   &bankAccountName,
				}

				payment, err := paymentDomain.CreatePayment(ctx, input)
				So(err, ShouldBeNil)
				So(payment, ShouldNotBeNil)
				So(payment.PaymentNumber, ShouldNotBeEmpty)
				So(payment.CustomerID, ShouldEqual, customer.ID)
				So(payment.Amount, ShouldEqual, 100000.0)
				So(payment.PaymentMethod, ShouldEqual, "bank_transfer")
				So(payment.Status, ShouldEqual, "pending")
				So(payment.AllocatedAmount, ShouldEqual, 0.0)

				Convey("GetPayment retrieves created payment", func() {
					found, err := paymentDomain.GetPayment(ctx, payment.ID.String())
					So(err, ShouldBeNil)
					So(found.ID, ShouldEqual, payment.ID)
					So(found.PaymentNumber, ShouldEqual, payment.PaymentNumber)
				})

				Convey("ListPayments returns payments", func() {
					filter := model.PaymentFilter{
						CustomerID: &customer.ID,
					}
					payments, err := paymentDomain.ListPayments(ctx, filter)
					So(err, ShouldBeNil)
					So(len(payments), ShouldBeGreaterThanOrEqualTo, 1)
				})

				Convey("UpdatePayment updates payment fields", func() {
					newBankName := "MANDIRI"
					notes := "Updated payment notes"
					input := model.PaymentInput{
						BankName: &newBankName,
						Notes:    &notes,
					}
					updated, err := paymentDomain.UpdatePayment(ctx, payment.ID.String(), input)
					So(err, ShouldBeNil)
					So(updated.BankName, ShouldNotBeNil)
					So(*updated.BankName, ShouldEqual, "MANDIRI")
					So(updated.Notes, ShouldNotBeNil)
					So(*updated.Notes, ShouldEqual, "Updated payment notes")
				})

				Convey("UpdatePayment with non-existent ID returns error", func() {
					input := model.PaymentInput{
						BankName: func() *string { s := "BRI"; return &s }(),
					}
					_, err := paymentDomain.UpdatePayment(ctx, uuid.New().String(), input)
					So(err, ShouldNotBeNil)
				})
			})

			Convey("CreatePayment with e-wallet method", func() {
				paymentDate := time.Now()
				amount := 50000.0
				paymentMethod := "e-wallet"
				ewalletProvider := "gopay"
				ewalletNumber := "08123456789"

				input := model.PaymentInput{
					CustomerID:      customer.ID,
					Amount:          &amount,
					PaymentMethod:   &paymentMethod,
					PaymentDate:     &paymentDate,
					EwalletProvider: &ewalletProvider,
					EwalletNumber:   &ewalletNumber,
				}

				payment, err := paymentDomain.CreatePayment(ctx, input)
				So(err, ShouldBeNil)
				So(payment.PaymentMethod, ShouldEqual, "e-wallet")
				So(payment.EwalletProvider, ShouldNotBeNil)
				So(*payment.EwalletProvider, ShouldEqual, "gopay")
			})

			Convey("CreatePayment with non-existent customer returns error", func() {
				paymentDate := time.Now()
				amount := 100000.0
				paymentMethod := "bank_transfer"

				input := model.PaymentInput{
					CustomerID:    uuid.New(),
					Amount:        &amount,
					PaymentMethod: &paymentMethod,
					PaymentDate:   &paymentDate,
				}

				_, err := paymentDomain.CreatePayment(ctx, input)
				So(err, ShouldNotBeNil)
			})

			Convey("DeletePayment removes payment", func() {
				// First create a payment
				paymentDate := time.Now()
				amount := 100000.0
				paymentMethod := "cash"

				input := model.PaymentInput{
					CustomerID:    customer.ID,
					Amount:        &amount,
					PaymentMethod: &paymentMethod,
					PaymentDate:   &paymentDate,
				}

				payment, err := paymentDomain.CreatePayment(ctx, input)
				So(err, ShouldBeNil)

				// Delete the payment
				err = paymentDomain.DeletePayment(ctx, payment.ID.String())
				So(err, ShouldBeNil)

				// Verify it's deleted
				_, err = paymentDomain.GetPayment(ctx, payment.ID.String())
				So(err, ShouldNotBeNil)
			})

			Convey("DeletePayment with empty ID returns error", func() {
				err := paymentDomain.DeletePayment(ctx, "")
				So(err, ShouldNotBeNil)
			})

			Convey("GetPayment with empty ID returns error", func() {
				_, err := paymentDomain.GetPayment(ctx, "")
				So(err, ShouldNotBeNil)
			})
		})

		Convey("Payment Allocation", func() {
			// Setup test data
			isActive := true
			profile := &model.BandwidthProfile{
				Name:         "Allocation Test Profile",
				Category:     "pppoe",
				PriceMonthly: 100000,
				TaxRate:      0.11,
				IsActive:     &isActive,
			}
			err := dbAdapter.BandwidthProfile().Create(profile)
			So(err, ShouldBeNil)

			customerCode := "ALLOC-" + time.Now().Format("20060102150405")
			expiryDate := time.Now().AddDate(0, 1, 0)
			customer := &model.Customer{
				CustomerCode: customerCode,
				FullName:     "Allocation Test Customer",
				Phone:        "08123456789",
				Status:       "active",
				ExpiryDate:   &expiryDate,
				ProfileID:    &profile.ID,
			}
			err = dbAdapter.Customer().Create(customer)
			So(err, ShouldBeNil)

			// Create invoice
			billingStart := time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC)
			billingEnd := time.Date(2026, time.February, 28, 23, 59, 59, 0, time.UTC)
			issueDate := time.Now()
			dueDate := issueDate.AddDate(0, 0, 7)

			// Create invoice items first
			item := &model.InvoiceItem{
				ItemType:  "subscription",
				ProfileID: &profile.ID,
				Quantity:  1,
				UnitPrice: 100000.0,
				Subtotal:  100000.0,
				TaxRate:   0.11,
				TaxAmount: 11000.0,
				Total:     111000.0,
			}
			err = dbAdapter.InvoiceItem().Create(item)
			So(err, ShouldBeNil)

			invoice := &model.Invoice{
				CustomerID:         customer.ID,
				BillingPeriodStart: billingStart,
				BillingPeriodEnd:   billingEnd,
				IssueDate:          issueDate,
				DueDate:            dueDate,
				Subtotal:           100000.0,
				TaxAmount:          11000.0,
				TotalAmount:        111000.0,
				Status:             "sent",
				PaymentStatus:      "unpaid",
			}
			model.InvoicePrepare(invoice)
			err = dbAdapter.Invoice().Create(invoice)
			So(err, ShouldBeNil)

			Convey("AllocatePayment allocates amount to invoice", func() {
				// Create payment
				paymentDate := time.Now()
				amount := 111000.0
				paymentMethod := "bank_transfer"

				paymentInput := model.PaymentInput{
					CustomerID:    customer.ID,
					InvoiceID:     &invoice.ID,
					Amount:        &amount,
					PaymentMethod: &paymentMethod,
					PaymentDate:   &paymentDate,
				}

				payment, err := paymentDomain.CreatePayment(ctx, paymentInput)
				So(err, ShouldBeNil)

				// Allocate payment to invoice
				allocationAmount := 50000.0
				err = paymentDomain.AllocatePayment(ctx, payment.ID.String(), invoice.ID.String(), allocationAmount)
				So(err, ShouldBeNil)

				// Verify allocation
				updatedPayment, err := paymentDomain.GetPayment(ctx, payment.ID.String())
				So(err, ShouldBeNil)
				So(updatedPayment.AllocatedAmount, ShouldEqual, 50000.0)

				Convey("AllocatePayment with zero amount returns error", func() {
					err := paymentDomain.AllocatePayment(ctx, payment.ID.String(), invoice.ID.String(), 0)
					So(err, ShouldNotBeNil)
				})

				Convey("AllocatePayment exceeding payment amount returns error", func() {
					err := paymentDomain.AllocatePayment(ctx, payment.ID.String(), invoice.ID.String(), 100000.0)
					So(err, ShouldNotBeNil)
				})

				Convey("AllocatePayment with invalid IDs returns error", func() {
					err := paymentDomain.AllocatePayment(ctx, uuid.New().String(), invoice.ID.String(), 10000.0)
					So(err, ShouldNotBeNil)
				})
			})

			Convey("GetPaymentHistory returns payment history", func() {
				// Create multiple payments
				for i := 0; i < 3; i++ {
					paymentDate := time.Now().Add(time.Duration(-i) * time.Hour * 24)
					amount := 50000.0
					paymentMethod := "cash"

					input := model.PaymentInput{
						CustomerID:    customer.ID,
						Amount:        &amount,
						PaymentMethod: &paymentMethod,
						PaymentDate:   &paymentDate,
					}

					_, err := paymentDomain.CreatePayment(ctx, input)
					So(err, ShouldBeNil)
				}

				filter := model.PaymentFilter{}
				payments, total, err := paymentDomain.GetPaymentHistory(ctx, customer.ID.String(), filter)
				So(err, ShouldBeNil)
				So(total, ShouldBeGreaterThanOrEqualTo, int64(3))
				So(len(payments), ShouldBeGreaterThanOrEqualTo, 3)
			})

			Convey("GetPaymentStatistics returns statistics", func() {
				// Create confirmed payments
				for i := 0; i < 2; i++ {
					paymentDate := time.Now()
					amount := 75000.0
					paymentMethod := "cash"

					input := model.PaymentInput{
						CustomerID:    customer.ID,
						Amount:        &amount,
						PaymentMethod: &paymentMethod,
						PaymentDate:   &paymentDate,
					}

					payment, err := paymentDomain.CreatePayment(ctx, input)
					So(err, ShouldBeNil)

					// Update to confirmed
					payment.Status = "confirmed"
					err = dbAdapter.Payment().Update(payment)
					So(err, ShouldBeNil)
				}

				stats, err := paymentDomain.GetPaymentStatistics(ctx, customer.ID.String())
				So(err, ShouldBeNil)
				So(stats.TotalPayments, ShouldBeGreaterThanOrEqualTo, int64(2))
				So(stats.TotalAmount, ShouldBeGreaterThan, 0)
				So(stats.AverageAmount, ShouldBeGreaterThan, 0)
			})
		})

		Convey("Webhook Processing", func() {
			Convey("ProcessWebhook with nil data returns error", func() {
				err := paymentDomain.ProcessWebhook(ctx, nil)
				So(err, ShouldNotBeNil)
			})

			Convey("ProcessWebhook with paid status updates payment", func() {
				// Create test data
				isActive := true
				profile := &model.BandwidthProfile{
					Name:         "Webhook Test Profile",
					Category:     "pppoe",
					PriceMonthly: 100000,
					TaxRate:      0.11,
					IsActive:     &isActive,
				}
				err := dbAdapter.BandwidthProfile().Create(profile)
				So(err, ShouldBeNil)

				customerCode := "WH-" + time.Now().Format("20060102150405")
				expiryDate := time.Now().AddDate(0, 1, 0)
				customer := &model.Customer{
					CustomerCode: customerCode,
					FullName:     "Webhook Test Customer",
					Phone:        "08123456789",
					Status:       "active",
					ExpiryDate:   &expiryDate,
					ProfileID:    &profile.ID,
				}
				err = dbAdapter.Customer().Create(customer)
				So(err, ShouldBeNil)

				// Create payment with xendit external ID
				paymentDate := time.Now()
				amount := 100000.0
				paymentMethod := "va"
				externalID := "test-external-id-" + time.Now().Format("20060102150405")

				input := model.PaymentInput{
					CustomerID:    customer.ID,
					Amount:        &amount,
					PaymentMethod: &paymentMethod,
					PaymentDate:   &paymentDate,
				}

				payment, err := paymentDomain.CreatePayment(ctx, input)
				So(err, ShouldBeNil)

				// Manually set xendit external ID (simulating xendit invoice creation)
				payment.XenditExternalID = &externalID
				payment.XenditInvoiceID = func() *string { s := "inv-test"; return &s }()
				err = dbAdapter.Payment().Update(payment)
				So(err, ShouldBeNil)

				// Process webhook
				webhookData := &xendit.WebhookData{
					ExternalID:    externalID,
					Status:        "PAID",
					PaymentID:     "pay-test-123",
					PaymentMethod: "VA",
					PaidAmount:    100000.0,
				}

				err = paymentDomain.ProcessWebhook(ctx, webhookData)
				So(err, ShouldBeNil)

				// Verify payment is updated
				updated, err := paymentDomain.GetPayment(ctx, payment.ID.String())
				So(err, ShouldBeNil)
				So(updated.Status, ShouldEqual, "confirmed")
				So(updated.AllocatedAmount, ShouldEqual, 100000.0)
			})
		})
	})
}
