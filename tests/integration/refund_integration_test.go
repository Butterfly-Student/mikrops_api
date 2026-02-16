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

func TestRefundIntegration(t *testing.T) {
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
		&model.Refund{},
		&model.Payment{},
		&model.Invoice{},
		&model.Customer{},
		&model.BandwidthProfile{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate refund tables: %v", err)
	}

	dbAdapter := postgres_outbound_adapter.NewAdapter(pgContainer.DB)
	refundDomain := domain.NewRefundDomain(dbAdapter)

	Convey("Test Refund Integration with PostgreSQL", t, func() {
		// Cleanup before test
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.Refund{})
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.Payment{})
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.Invoice{})
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.Customer{})
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.BandwidthProfile{})

		Convey("Setup test data", func() {
			// Create bandwidth profile
			profile := &model.BandwidthProfile{
				Name:         "Test Refund Profile",
				Category:     "pppoe",
				PriceMonthly: 100000,
				TaxRate:      0.11,
				IsActive:     true,
			}
			err := dbAdapter.BandwidthProfile().Create(profile)
			So(err, ShouldBeNil)

			// Create customer
			customerCode := "REF-" + time.Now().Format("20060102150405")
			email := "refund@test.com"
			expiryDate := time.Now().AddDate(0, 1, 0)
			customer := &model.Customer{
				CustomerCode: customerCode,
				FullName:     "Refund Test Customer",
				Email:        &email,
				Phone:        "08123456789",
				Status:       "active",
				ExpiryDate:   &expiryDate,
				ProfileID:    &profile.ID,
			}
			err = dbAdapter.Customer().Create(customer)
			So(err, ShouldBeNil)

			// Create payment
			paymentDate := time.Now()
			amount := 111000.0
			paymentMethod := "bank_transfer"

			payment := &model.Payment{
				CustomerID:    customer.ID,
				Amount:        amount,
				PaymentMethod: paymentMethod,
				PaymentDate:   paymentDate,
				Status:        "confirmed",
			}
			model.PaymentPrepare(payment)
			err = dbAdapter.Payment().Create(payment)
			So(err, ShouldBeNil)

			Convey("CreateRefund creates refund successfully", func() {
				refundAmount := 50000.0
				refundType := "partial"
				refundReason := "Customer request - service not used"
				refundMethod := "bank_transfer"
				bankName := "BCA"
				bankAccountName := "Test Account"
				bankAccountNumber := "1234567890"

				input := model.RefundInput{
					PaymentID:         payment.ID,
					RefundAmount:      &refundAmount,
					RefundType:        &refundType,
					RefundReason:      &refundReason,
					RefundMethod:      &refundMethod,
					BankName:          &bankName,
					BankAccountName:   &bankAccountName,
					BankAccountNumber: &bankAccountNumber,
				}

				refund, err := refundDomain.Create(ctx, input)
				So(err, ShouldBeNil)
				So(refund, ShouldNotBeNil)
				So(refund.RefundNumber, ShouldNotBeEmpty)
				So(refund.PaymentID, ShouldEqual, payment.ID)
				So(refund.CustomerID, ShouldEqual, customer.ID)
				So(refund.RefundAmount, ShouldEqual, 50000.0)
				So(refund.RefundType, ShouldEqual, "partial")
				So(refund.RefundReason, ShouldEqual, "Customer request - service not used")
				So(refund.RefundMethod, ShouldEqual, "bank_transfer")
				So(refund.Status, ShouldEqual, "pending")

				Convey("GetRefund retrieves created refund", func() {
					found, err := refundDomain.Get(ctx, refund.ID.String())
					So(err, ShouldBeNil)
					So(found.ID, ShouldEqual, refund.ID)
					So(found.RefundNumber, ShouldEqual, refund.RefundNumber)
				})

				Convey("ListRefunds returns refunds", func() {
					filter := model.RefundFilter{}
					refunds, err := refundDomain.List(ctx, filter)
					So(err, ShouldBeNil)
					So(len(refunds), ShouldBeGreaterThanOrEqualTo, 1)
				})

				Convey("ListRefunds with customer filter", func() {
					filter := model.RefundFilter{
						CustomerIDs: []uuid.UUID{customer.ID},
					}
					refunds, err := refundDomain.List(ctx, filter)
					So(err, ShouldBeNil)
					So(len(refunds), ShouldBeGreaterThanOrEqualTo, 1)
				})

				Convey("ListRefunds with status filter", func() {
					filter := model.RefundFilter{
						Status: []string{"pending"},
					}
					refunds, err := refundDomain.List(ctx, filter)
					So(err, ShouldBeNil)
					So(len(refunds), ShouldBeGreaterThanOrEqualTo, 1)
				})

				Convey("ApproveRefund approves refund", func() {
					approvedBy := uuid.New()

					err := refundDomain.Approve(ctx, refund.ID.String(), approvedBy)
					So(err, ShouldBeNil)

					updated, err := refundDomain.Get(ctx, refund.ID.String())
					So(err, ShouldBeNil)
					So(updated.Status, ShouldEqual, "approved")
					So(updated.ApprovedBy, ShouldNotBeNil)
					So(*updated.ApprovedBy, ShouldEqual, approvedBy)
					So(updated.ApprovedAt, ShouldNotBeNil)
				})

				Convey("RejectRefund rejects refund", func() {
					rejectedBy := uuid.New()
					rejectionReason := "Payment already allocated to invoice"

					err := refundDomain.Reject(ctx, refund.ID.String(), rejectedBy, rejectionReason)
					So(err, ShouldBeNil)

					updated, err := refundDomain.Get(ctx, refund.ID.String())
					So(err, ShouldBeNil)
					So(updated.Status, ShouldEqual, "rejected")
					So(updated.RejectionReason, ShouldNotBeNil)
					So(*updated.RejectionReason, ShouldEqual, "Payment already allocated to invoice")
				})

				Convey("ProcessRefund processes refund", func() {
					// First approve
					approvedBy := uuid.New()
					err := refundDomain.Approve(ctx, refund.ID.String(), approvedBy)
					So(err, ShouldBeNil)

					// Then process
					processedBy := uuid.New()

					err = refundDomain.Process(ctx, refund.ID.String(), processedBy)
					So(err, ShouldBeNil)

					updated, err := refundDomain.Get(ctx, refund.ID.String())
					So(err, ShouldBeNil)
					So(updated.Status, ShouldEqual, "processed")
					So(updated.ProcessedBy, ShouldNotBeNil)
					So(*updated.ProcessedBy, ShouldEqual, processedBy)
					So(updated.ProcessedAt, ShouldNotBeNil)
				})

				Convey("CompleteRefund completes refund", func() {
					// First approve and process
					approvedBy := uuid.New()
					err := refundDomain.Approve(ctx, refund.ID.String(), approvedBy)
					So(err, ShouldBeNil)

					processedBy := uuid.New()
					err = refundDomain.Process(ctx, refund.ID.String(), processedBy)
					So(err, ShouldBeNil)

					// Then complete
					err = refundDomain.Complete(ctx, refund.ID.String())
					So(err, ShouldBeNil)

					updated, err := refundDomain.Get(ctx, refund.ID.String())
					So(err, ShouldBeNil)
					So(updated.Status, ShouldEqual, "completed")
				})

				Convey("UpdateRefund updates refund fields", func() {
					newRefundAmount := 30000.0
					notes := "Updated refund details"

					input := model.RefundInput{
						PaymentID:    payment.ID,
						RefundAmount: &newRefundAmount,
						Notes:        &notes,
					}

					updated, err := refundDomain.Update(ctx, refund.ID.String(), input)
					So(err, ShouldBeNil)
					So(updated.RefundAmount, ShouldEqual, 30000.0)
					So(updated.Notes, ShouldNotBeNil)
					So(*updated.Notes, ShouldEqual, "Updated refund details")
				})

				Convey("DeleteRefund soft deletes refund", func() {
					err := refundDomain.Delete(ctx, refund.ID.String())
					So(err, ShouldBeNil)

					// Verify soft delete
					_, err = refundDomain.Get(ctx, refund.ID.String())
					So(err, ShouldNotBeNil)
				})
			})

			Convey("CreateRefund with full refund type", func() {
				refundAmount := 111000.0
				refundType := "full"
				refundReason := "Service cancellation"
				refundMethod := "original"

				input := model.RefundInput{
					PaymentID:    payment.ID,
					RefundAmount: &refundAmount,
					RefundType:   &refundType,
					RefundReason: &refundReason,
					RefundMethod: &refundMethod,
				}

				refund, err := refundDomain.Create(ctx, input)
				So(err, ShouldBeNil)
				So(refund.RefundType, ShouldEqual, "full")
				So(refund.RefundAmount, ShouldEqual, 111000.0)
			})

			Convey("CreateRefund with cash method", func() {
				refundAmount := 50000.0
				refundType := "partial"
				refundReason := "Overpayment"
				refundMethod := "cash"

				input := model.RefundInput{
					PaymentID:    payment.ID,
					RefundAmount: &refundAmount,
					RefundType:   &refundType,
					RefundReason: &refundReason,
					RefundMethod: &refundMethod,
				}

				refund, err := refundDomain.Create(ctx, input)
				So(err, ShouldBeNil)
				So(refund.RefundMethod, ShouldEqual, "cash")
			})

			Convey("CreateRefund with invalid payment returns error", func() {
				refundAmount := 50000.0
				refundType := "partial"
				refundReason := "Test"
				refundMethod := "cash"

				input := model.RefundInput{
					PaymentID:    uuid.New(),
					RefundAmount: &refundAmount,
					RefundType:   &refundType,
					RefundReason: &refundReason,
					RefundMethod: &refundMethod,
				}

				_, err := refundDomain.Create(ctx, input)
				So(err, ShouldNotBeNil)
			})

			Convey("CreateRefund with zero amount returns error", func() {
				refundAmount := 0.0
				refundType := "partial"
				refundReason := "Test"
				refundMethod := "cash"

				input := model.RefundInput{
					PaymentID:    payment.ID,
					RefundAmount: &refundAmount,
					RefundType:   &refundType,
					RefundReason: &refundReason,
					RefundMethod: &refundMethod,
				}

				_, err := refundDomain.Create(ctx, input)
				So(err, ShouldNotBeNil)
			})

			Convey("GetRefund with invalid ID returns error", func() {
				_, err := refundDomain.Get(ctx, uuid.New().String())
				So(err, ShouldNotBeNil)
			})

			Convey("UpdateRefund with invalid ID returns error", func() {
				refundAmount := 10000.0
				input := model.RefundInput{
					PaymentID:    payment.ID,
					RefundAmount: &refundAmount,
				}
				_, err := refundDomain.Update(ctx, uuid.New().String(), input)
				So(err, ShouldNotBeNil)
			})

			Convey("DeleteRefund with invalid ID returns error", func() {
				err := refundDomain.Delete(ctx, uuid.New().String())
				So(err, ShouldNotBeNil)
			})

			Convey("ApproveRefund with invalid ID returns error", func() {
				approvedBy := uuid.New()
				err := refundDomain.Approve(ctx, uuid.New().String(), approvedBy)
				So(err, ShouldNotBeNil)
			})

			Convey("RejectRefund with invalid ID returns error", func() {
				rejectedBy := uuid.New()
				err := refundDomain.Reject(ctx, uuid.New().String(), rejectedBy, "Test reason")
				So(err, ShouldNotBeNil)
			})

			Convey("ProcessRefund with invalid ID returns error", func() {
				processedBy := uuid.New()
				err := refundDomain.Process(ctx, uuid.New().String(), processedBy)
				So(err, ShouldNotBeNil)
			})

			Convey("CompleteRefund with invalid ID returns error", func() {
				err := refundDomain.Complete(ctx, uuid.New().String())
				So(err, ShouldNotBeNil)
			})

			Convey("Approve already approved refund returns error", func() {
				refundAmount := 50000.0
				refundType := "partial"
				refundReason := "Test"
				refundMethod := "cash"

				input := model.RefundInput{
					PaymentID:    payment.ID,
					RefundAmount: &refundAmount,
					RefundType:   &refundType,
					RefundReason: &refundReason,
					RefundMethod: &refundMethod,
				}

				refund, err := refundDomain.Create(ctx, input)
				So(err, ShouldBeNil)

				approvedBy := uuid.New()
				err = refundDomain.Approve(ctx, refund.ID.String(), approvedBy)
				So(err, ShouldBeNil)

				// Try to approve again
				err = refundDomain.Approve(ctx, refund.ID.String(), approvedBy)
				So(err, ShouldNotBeNil)
			})

			Convey("Reject non-pending refund returns error", func() {
				refundAmount := 50000.0
				refundType := "partial"
				refundReason := "Test"
				refundMethod := "cash"

				input := model.RefundInput{
					PaymentID:    payment.ID,
					RefundAmount: &refundAmount,
					RefundType:   &refundType,
					RefundReason: &refundReason,
					RefundMethod: &refundMethod,
				}

				refund, err := refundDomain.Create(ctx, input)
				So(err, ShouldBeNil)

				// First approve
				approvedBy := uuid.New()
				err = refundDomain.Approve(ctx, refund.ID.String(), approvedBy)
				So(err, ShouldBeNil)

				// Try to reject approved refund
				rejectedBy := uuid.New()
				err = refundDomain.Reject(ctx, refund.ID.String(), rejectedBy, "Test reason")
				So(err, ShouldNotBeNil)
			})

			Convey("Process non-approved refund returns error", func() {
				refundAmount := 50000.0
				refundType := "partial"
				refundReason := "Test"
				refundMethod := "cash"

				input := model.RefundInput{
					PaymentID:    payment.ID,
					RefundAmount: &refundAmount,
					RefundType:   &refundType,
					RefundReason: &refundReason,
					RefundMethod: &refundMethod,
				}

				refund, err := refundDomain.Create(ctx, input)
				So(err, ShouldBeNil)

				// Try to process pending refund
				processedBy := uuid.New()
				err = refundDomain.Process(ctx, refund.ID.String(), processedBy)
				So(err, ShouldNotBeNil)
			})
		})

		Convey("Refund filtering", func() {
			// Create test data
			profile := &model.BandwidthProfile{
				Name:         "Filter Refund Profile",
				Category:     "pppoe",
				PriceMonthly: 100000,
				TaxRate:      0.11,
				IsActive:     true,
			}
			err := dbAdapter.BandwidthProfile().Create(profile)
			So(err, ShouldBeNil)

			customerCode := "FILTREF-" + time.Now().Format("20060102150405")
			email := "filterrefund@test.com"
			expiryDate := time.Now().AddDate(0, 1, 0)
			customer := &model.Customer{
				CustomerCode: customerCode,
				FullName:     "Filter Refund Customer",
				Email:        &email,
				Phone:        "08123456789",
				Status:       "active",
				ExpiryDate:   &expiryDate,
				ProfileID:    &profile.ID,
			}
			err = dbAdapter.Customer().Create(customer)
			So(err, ShouldBeNil)

			// Create multiple payments
			var payments []model.Payment
			for i := 0; i < 3; i++ {
				payment := &model.Payment{
					CustomerID:    customer.ID,
					Amount:        float64((i + 1) * 50000),
					PaymentMethod: "bank_transfer",
					PaymentDate:   time.Now(),
					Status:        "confirmed",
				}
				model.PaymentPrepare(payment)
				err = dbAdapter.Payment().Create(payment)
				So(err, ShouldBeNil)
				payments = append(payments, *payment)
			}

			// Create refunds
			for i := 0; i < 3; i++ {
				refundAmount := float64((i + 1) * 25000)
				refundType := "partial"
				refundReason := "Refund " + string(rune('1'+i))
				refundMethod := "bank_transfer"

				input := model.RefundInput{
					PaymentID:    payments[i].ID,
					RefundAmount: &refundAmount,
					RefundType:   &refundType,
					RefundReason: &refundReason,
					RefundMethod: &refundMethod,
				}

				_, err := refundDomain.Create(ctx, input)
				So(err, ShouldBeNil)
			}

			Convey("Filter by refund type", func() {
				filter := model.RefundFilter{
					RefundType: []string{"partial"},
				}
				refunds, err := refundDomain.List(ctx, filter)
				So(err, ShouldBeNil)
				So(len(refunds), ShouldBeGreaterThanOrEqualTo, 3)
			})

			Convey("Filter by refund method", func() {
				filter := model.RefundFilter{
					RefundMethod: []string{"bank_transfer"},
				}
				refunds, err := refundDomain.List(ctx, filter)
				So(err, ShouldBeNil)
				So(len(refunds), ShouldBeGreaterThanOrEqualTo, 3)
			})

			Convey("Filter by amount range", func() {
				minAmount := 25000.0
				maxAmount := 50000.0

				filter := model.RefundFilter{
					AmountMin: &minAmount,
					AmountMax: &maxAmount,
				}
				refunds, err := refundDomain.List(ctx, filter)
				So(err, ShouldBeNil)
				So(len(refunds), ShouldBeGreaterThanOrEqualTo, 2)
			})

			Convey("Filter by payment IDs", func() {
				filter := model.RefundFilter{
					PaymentIDs: []uuid.UUID{payments[0].ID, payments[1].ID},
				}
				refunds, err := refundDomain.List(ctx, filter)
				So(err, ShouldBeNil)
				So(len(refunds), ShouldBeGreaterThanOrEqualTo, 2)
			})
		})

		Convey("Refund statistics", func() {
			// Create test data
			profile := &model.BandwidthProfile{
				Name:         "Stats Refund Profile",
				Category:     "pppoe",
				PriceMonthly: 100000,
				TaxRate:      0.11,
				IsActive:     true,
			}
			err := dbAdapter.BandwidthProfile().Create(profile)
			So(err, ShouldBeNil)

			customerCode := "STATREF-" + time.Now().Format("20060102150405")
			email := "statsrefund@test.com"
			expiryDate := time.Now().AddDate(0, 1, 0)
			customer := &model.Customer{
				CustomerCode: customerCode,
				FullName:     "Stats Refund Customer",
				Email:        &email,
				Phone:        "08123456789",
				Status:       "active",
				ExpiryDate:   &expiryDate,
				ProfileID:    &profile.ID,
			}
			err = dbAdapter.Customer().Create(customer)
			So(err, ShouldBeNil)

			payment := &model.Payment{
				CustomerID:    customer.ID,
				Amount:        150000.0,
				PaymentMethod: "bank_transfer",
				PaymentDate:   time.Now(),
				Status:        "confirmed",
			}
			model.PaymentPrepare(payment)
			err = dbAdapter.Payment().Create(payment)
			So(err, ShouldBeNil)

			// Create refunds with different statuses
			for i := 0; i < 4; i++ {
				refundAmount := 25000.0
				refundType := "partial"
				refundReason := "Test refund"
				refundMethod := "cash"

				input := model.RefundInput{
					PaymentID:    payment.ID,
					RefundAmount: &refundAmount,
					RefundType:   &refundType,
					RefundReason: &refundReason,
					RefundMethod: &refundMethod,
				}

				refund, err := refundDomain.Create(ctx, input)
				So(err, ShouldBeNil)

				// Change status
				if i == 1 {
					approvedBy := uuid.New()
					err = refundDomain.Approve(ctx, refund.ID.String(), approvedBy)
					So(err, ShouldBeNil)
				} else if i == 2 {
					approvedBy := uuid.New()
					err = refundDomain.Approve(ctx, refund.ID.String(), approvedBy)
					So(err, ShouldBeNil)

					processedBy := uuid.New()
					err = refundDomain.Process(ctx, refund.ID.String(), processedBy)
					So(err, ShouldBeNil)
				} else if i == 3 {
					rejectedBy := uuid.New()
					err = refundDomain.Reject(ctx, refund.ID.String(), rejectedBy, "Test")
					So(err, ShouldBeNil)
				}
			}

			Convey("Count refunds by status", func() {
				filter := model.RefundFilter{
					Status: []string{"pending"},
				}
				refunds, err := refundDomain.List(ctx, filter)
				So(err, ShouldBeNil)
				So(len(refunds), ShouldBeGreaterThanOrEqualTo, 1)

				filter = model.RefundFilter{
					Status: []string{"approved"},
				}
				refunds, err = refundDomain.List(ctx, filter)
				So(err, ShouldBeNil)
				So(len(refunds), ShouldBeGreaterThanOrEqualTo, 1)

				filter = model.RefundFilter{
					Status: []string{"rejected"},
				}
				refunds, err = refundDomain.List(ctx, filter)
				So(err, ShouldBeNil)
				So(len(refunds), ShouldBeGreaterThanOrEqualTo, 1)
			})

			Convey("Calculate total refund amount", func() {
				filter := model.RefundFilter{
					CustomerIDs: []uuid.UUID{customer.ID},
				}
				refunds, err := refundDomain.List(ctx, filter)
				So(err, ShouldBeNil)

				totalAmount := 0.0
				for _, refund := range refunds {
					totalAmount += refund.RefundAmount
				}
				So(totalAmount, ShouldBeGreaterThan, 0)
			})
		})
	})
}
