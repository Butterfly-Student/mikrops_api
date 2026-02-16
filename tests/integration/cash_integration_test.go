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

func TestCashIntegration(t *testing.T) {
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
		&model.CashTransaction{},
		&model.CashCategory{},
		&model.Customer{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate cash tables: %v", err)
	}

	dbAdapter := postgres_outbound_adapter.NewAdapter(pgContainer.DB)
	cashDomain := domain.NewCashDomain(dbAdapter)

	Convey("Test Cash Integration with PostgreSQL", t, func() {
		// Cleanup before test
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.CashTransaction{})
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.CashCategory{})
		pgContainer.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.Customer{})

		Convey("Setup test data - Cash Categories", func() {
			Convey("CreateCashCategory creates category successfully", func() {
				code := "CAT-" + time.Now().Format("20060102150405")
				name := "Service Income"
				categoryType := "income"
				description := "Income from internet services"

				input := model.CashCategoryInput{
					Code:        &code,
					Name:        &name,
					Type:        &categoryType,
					Description: &description,
					IsActive:    func() *bool { b := true; return &b }(),
					SortOrder:   func() *int { i := 1; return &i }(),
				}

				category, err := cashDomain.CreateCategory(ctx, input)
				So(err, ShouldBeNil)
				So(category, ShouldNotBeNil)
				So(category.Code, ShouldEqual, code)
				So(category.Name, ShouldEqual, "Service Income")
				So(category.Type, ShouldEqual, "income")
				So(category.Description, ShouldNotBeNil)
				So(*category.Description, ShouldEqual, "Income from internet services")
				So(category.IsActive, ShouldEqual, true)
				So(category.SortOrder, ShouldEqual, 1)

				Convey("GetCategory retrieves created category", func() {
					found, err := cashDomain.GetCategory(ctx, category.ID.String())
					So(err, ShouldBeNil)
					So(found.ID, ShouldEqual, category.ID)
					So(found.Code, ShouldEqual, code)
				})

				Convey("ListCategories returns categories", func() {
					filter := model.CashCategoryFilter{}
					categories, err := cashDomain.ListCategories(ctx, filter)
					So(err, ShouldBeNil)
					So(len(categories), ShouldBeGreaterThanOrEqualTo, 1)
				})

				Convey("ListCategories with type filter", func() {
					filter := model.CashCategoryFilter{
						Type: &categoryType,
					}
					categories, err := cashDomain.ListCategories(ctx, filter)
					So(err, ShouldBeNil)
					So(len(categories), ShouldBeGreaterThanOrEqualTo, 1)
					for _, c := range categories {
						So(c.Type, ShouldEqual, "income")
					}
				})

				Convey("UpdateCategory updates category", func() {
					newName := "Updated Service Income"
					newDescription := "Updated description"

					input := model.CashCategoryInput{
						Name:        &newName,
						Description: &newDescription,
					}

					updated, err := cashDomain.UpdateCategory(ctx, category.ID.String(), input)
					So(err, ShouldBeNil)
					So(updated.Name, ShouldEqual, "Updated Service Income")
					So(updated.Description, ShouldNotBeNil)
					So(*updated.Description, ShouldEqual, "Updated description")
				})

				Convey("DeleteCategory removes category", func() {
					err := cashDomain.DeleteCategory(ctx, category.ID.String())
					So(err, ShouldBeNil)

					// Verify it's deleted
					_, err = cashDomain.GetCategory(ctx, category.ID.String())
					So(err, ShouldNotBeNil)
				})
			})

			Convey("Create expense category", func() {
				code := "EXP-" + time.Now().Format("20060102150405")
				name := "Office Expenses"
				categoryType := "expense"

				input := model.CashCategoryInput{
					Code:     &code,
					Name:     &name,
					Type:     &categoryType,
					IsActive: func() *bool { b := true; return &b }(),
				}

				category, err := cashDomain.CreateCategory(ctx, input)
				So(err, ShouldBeNil)
				So(category.Type, ShouldEqual, "expense")
			})

			Convey("Create duplicate category code returns error", func() {
				code := "DUP-" + time.Now().Format("20060102150405")
				name := "Duplicate"
				categoryType := "income"

				input := model.CashCategoryInput{
					Code:     &code,
					Name:     &name,
					Type:     &categoryType,
					IsActive: func() *bool { b := true; return &b }(),
				}

				_, err := cashDomain.CreateCategory(ctx, input)
				So(err, ShouldBeNil)

				// Try to create duplicate
				input2 := model.CashCategoryInput{
					Code:     &code,
					Name:     func() *string { s := "Duplicate 2"; return &s }(),
					Type:     &categoryType,
					IsActive: func() *bool { b := true; return &b }(),
				}

				_, err = cashDomain.CreateCategory(ctx, input2)
				So(err, ShouldNotBeNil)
			})
		})

		Convey("Setup test data - Customer", func() {
			profile := &model.BandwidthProfile{
				Name:         "Cash Test Profile",
				Category:     "pppoe",
				PriceMonthly: 100000,
				TaxRate:      0.11,
				IsActive:     true,
			}
			err := pgContainer.DB.Create(profile).Error
			So(err, ShouldBeNil)

			customerCode := "CASH-" + time.Now().Format("20060102150405")
			email := "cash@test.com"
			expiryDate := time.Now().AddDate(0, 1, 0)
			customer := &model.Customer{
				CustomerCode: customerCode,
				FullName:     "Cash Test Customer",
				Email:        &email,
				Phone:        "08123456789",
				Status:       "active",
				ExpiryDate:   &expiryDate,
				ProfileID:    &profile.ID,
			}
			err = pgContainer.DB.Create(customer).Error
			So(err, ShouldBeNil)

			Convey("CreateCashTransaction creates transaction successfully", func() {
				// Create category first
				code := "SRV-" + time.Now().Format("20060102150405")
				name := "Service Income"
				categoryType := "income"

				categoryInput := model.CashCategoryInput{
					Code:     &code,
					Name:     &name,
					Type:     &categoryType,
					IsActive: func() *bool { b := true; return &b }(),
				}

				category, err := cashDomain.CreateCategory(ctx, categoryInput)
				So(err, ShouldBeNil)

				// Create transaction
				transactionDate := time.Now()
				trxType := "income"
				amount := 100000.0
				paymentMethod := "cash"
				description := "Monthly payment from customer"
				referenceType := "payment"

				input := model.CashTransactionInput{
					TransactionDate: &transactionDate,
					Type:            &trxType,
					CategoryID:      category.ID,
					Amount:          &amount,
					PaymentMethod:   &paymentMethod,
					Description:     &description,
					ReferenceType:   &referenceType,
					CustomerID:      &customer.ID,
				}

				transaction, err := cashDomain.CreateTransaction(ctx, input)
				So(err, ShouldBeNil)
				So(transaction, ShouldNotBeNil)
				So(transaction.TransactionNumber, ShouldNotBeEmpty)
				So(transaction.Type, ShouldEqual, "income")
				So(transaction.CategoryID, ShouldEqual, category.ID)
				So(transaction.Amount, ShouldEqual, 100000.0)
				So(transaction.PaymentMethod, ShouldNotBeNil)
				So(*transaction.PaymentMethod, ShouldEqual, "cash")
				So(transaction.Description, ShouldNotBeNil)
				So(*transaction.Description, ShouldEqual, "Monthly payment from customer")
				So(transaction.CustomerID, ShouldNotBeNil)
				So(*transaction.CustomerID, ShouldEqual, customer.ID)
				So(transaction.ApprovalStatus, ShouldEqual, "pending")

				Convey("GetTransaction retrieves created transaction", func() {
					found, err := cashDomain.GetTransaction(ctx, transaction.ID.String())
					So(err, ShouldBeNil)
					So(found.ID, ShouldEqual, transaction.ID)
					So(found.TransactionNumber, ShouldEqual, transaction.TransactionNumber)
				})

				Convey("ListTransactions returns transactions", func() {
					filter := model.CashTransactionFilter{}
					transactions, err := cashDomain.ListTransactions(ctx, filter)
					So(err, ShouldBeNil)
					So(len(transactions), ShouldBeGreaterThanOrEqualTo, 1)
				})

				Convey("ListTransactions with type filter", func() {
					filter := model.CashTransactionFilter{
						Type: &trxType,
					}
					transactions, err := cashDomain.ListTransactions(ctx, filter)
					So(err, ShouldBeNil)
					So(len(transactions), ShouldBeGreaterThanOrEqualTo, 1)
					for _, t := range transactions {
						So(t.Type, ShouldEqual, "income")
					}
				})

				Convey("ListTransactions with category filter", func() {
					filter := model.CashTransactionFilter{
						CategoryIDs: []uuid.UUID{category.ID},
					}
					transactions, err := cashDomain.ListTransactions(ctx, filter)
					So(err, ShouldBeNil)
					So(len(transactions), ShouldBeGreaterThanOrEqualTo, 1)
				})

				Convey("ListTransactions with customer filter", func() {
					filter := model.CashTransactionFilter{
						CustomerIDs: []uuid.UUID{customer.ID},
					}
					transactions, err := cashDomain.ListTransactions(ctx, filter)
					So(err, ShouldBeNil)
					So(len(transactions), ShouldBeGreaterThanOrEqualTo, 1)
				})

				Convey("ListTransactions with date range filter", func() {
					startDate := time.Now().Add(-time.Hour * 24)
					endDate := time.Now().Add(time.Hour * 24)

					filter := model.CashTransactionFilter{
						DateStart: &startDate,
						DateEnd:   &endDate,
					}
					transactions, err := cashDomain.ListTransactions(ctx, filter)
					So(err, ShouldBeNil)
					So(len(transactions), ShouldBeGreaterThanOrEqualTo, 1)
				})

				Convey("UpdateTransaction updates transaction", func() {
					newAmount := 150000.0
					newDescription := "Updated payment description"
					notes := "Updated notes"

					input := model.CashTransactionInput{
						CategoryID:  transaction.CategoryID,
						Amount:      &newAmount,
						Description: &newDescription,
						Notes:       &notes,
					}

					updated, err := cashDomain.UpdateTransaction(ctx, transaction.ID.String(), input)
					So(err, ShouldBeNil)
					So(updated.Amount, ShouldEqual, 150000.0)
					So(updated.Description, ShouldNotBeNil)
					So(*updated.Description, ShouldEqual, "Updated payment description")
					So(updated.Notes, ShouldNotBeNil)
					So(*updated.Notes, ShouldEqual, "Updated notes")
				})

				Convey("DeleteTransaction removes transaction", func() {
					err := cashDomain.DeleteTransaction(ctx, transaction.ID.String())
					So(err, ShouldBeNil)

					// Verify it's deleted
					_, err = cashDomain.GetTransaction(ctx, transaction.ID.String())
					So(err, ShouldNotBeNil)
				})
			})

			Convey("CreateCashTransaction with expense type", func() {
				code := "OFF-" + time.Now().Format("20060102150405")
				name := "Office Expenses"
				categoryType := "expense"

				categoryInput := model.CashCategoryInput{
					Code:     &code,
					Name:     &name,
					Type:     &categoryType,
					IsActive: func() *bool { b := true; return &b }(),
				}

				category, err := cashDomain.CreateCategory(ctx, categoryInput)
				So(err, ShouldBeNil)

				transactionDate := time.Now()
				trxType := "expense"
				amount := 50000.0
				paymentMethod := "bank_transfer"
				description := "Office supplies purchase"

				input := model.CashTransactionInput{
					TransactionDate: &transactionDate,
					Type:            &trxType,
					CategoryID:      category.ID,
					Amount:          &amount,
					PaymentMethod:   &paymentMethod,
					Description:     &description,
				}

				transaction, err := cashDomain.CreateTransaction(ctx, input)
				So(err, ShouldBeNil)
				So(transaction.Type, ShouldEqual, "expense")
				So(transaction.Amount, ShouldEqual, 50000.0)
			})

			Convey("CreateCashTransaction with approval required", func() {
				code := "APPR-" + time.Now().Format("20060102150405")
				name := "Large Expenses"
				categoryType := "expense"

				categoryInput := model.CashCategoryInput{
					Code:     &code,
					Name:     &name,
					Type:     &categoryType,
					IsActive: func() *bool { b := true; return &b }(),
				}

				category, err := cashDomain.CreateCategory(ctx, categoryInput)
				So(err, ShouldBeNil)

				transactionDate := time.Now()
				trxType := "expense"
				amount := 1000000.0
				requiresApproval := true

				input := model.CashTransactionInput{
					TransactionDate:  &transactionDate,
					Type:              &trxType,
					CategoryID:        category.ID,
					Amount:            &amount,
					Description:       func() *string { s := "Large equipment purchase"; return &s }(),
					RequiresApproval:  &requiresApproval,
				}

				transaction, err := cashDomain.CreateTransaction(ctx, input)
				So(err, ShouldBeNil)
				So(transaction.RequiresApproval, ShouldEqual, true)
				So(transaction.ApprovalStatus, ShouldEqual, "pending")

				Convey("ApproveTransaction approves transaction", func() {
					approvedBy := uuid.New()

					err := cashDomain.ApproveTransaction(ctx, transaction.ID.String(), approvedBy)
					So(err, ShouldBeNil)

					updated, err := cashDomain.GetTransaction(ctx, transaction.ID.String())
					So(err, ShouldBeNil)
					So(updated.ApprovalStatus, ShouldEqual, "approved")
					So(updated.ApprovedBy, ShouldNotBeNil)
					So(*updated.ApprovedBy, ShouldEqual, approvedBy)
					So(updated.ApprovedAt, ShouldNotBeNil)
				})

				Convey("RejectTransaction rejects transaction", func() {
					rejectedBy := uuid.New()

					err := cashDomain.RejectTransaction(ctx, transaction.ID.String(), rejectedBy)
					So(err, ShouldBeNil)

					updated, err := cashDomain.GetTransaction(ctx, transaction.ID.String())
					So(err, ShouldBeNil)
					So(updated.ApprovalStatus, ShouldEqual, "rejected")
					So(updated.ApprovedBy, ShouldNotBeNil)
					So(*updated.ApprovedBy, ShouldEqual, rejectedBy)
					So(updated.ApprovedAt, ShouldNotBeNil)
				})
			})
		})

		Convey("Error handling", func() {
			Convey("GetCategory with invalid ID returns error", func() {
				_, err := cashDomain.GetCategory(ctx, uuid.New().String())
				So(err, ShouldNotBeNil)
			})

			Convey("UpdateCategory with invalid ID returns error", func() {
				input := model.CashCategoryInput{
					Name: func() *string { s := "Test"; return &s }(),
				}
				_, err := cashDomain.UpdateCategory(ctx, uuid.New().String(), input)
				So(err, ShouldNotBeNil)
			})

			Convey("DeleteCategory with invalid ID returns error", func() {
				err := cashDomain.DeleteCategory(ctx, uuid.New().String())
				So(err, ShouldNotBeNil)
			})

			Convey("GetTransaction with invalid ID returns error", func() {
				_, err := cashDomain.GetTransaction(ctx, uuid.New().String())
				So(err, ShouldNotBeNil)
			})

			Convey("UpdateTransaction with invalid ID returns error", func() {
				// Create a category first
				code := "ERR-" + time.Now().Format("20060102150405")
				name := "Error Test"
				categoryType := "income"

				categoryInput := model.CashCategoryInput{
					Code:     &code,
					Name:     &name,
					Type:     &categoryType,
					IsActive: func() *bool { b := true; return &b }(),
				}

				category, err := cashDomain.CreateCategory(ctx, categoryInput)
				So(err, ShouldBeNil)

				input := model.CashTransactionInput{
					CategoryID: category.ID,
					Amount:     func() *float64 { f := 100.0; return &f }(),
				}
				_, err = cashDomain.UpdateTransaction(ctx, uuid.New().String(), input)
				So(err, ShouldNotBeNil)
			})

			Convey("DeleteTransaction with invalid ID returns error", func() {
				err := cashDomain.DeleteTransaction(ctx, uuid.New().String())
				So(err, ShouldNotBeNil)
			})

			Convey("ApproveTransaction with invalid ID returns error", func() {
				approvedBy := uuid.New()
				err := cashDomain.ApproveTransaction(ctx, uuid.New().String(), approvedBy)
				So(err, ShouldNotBeNil)
			})

			Convey("RejectTransaction with invalid ID returns error", func() {
				rejectedBy := uuid.New()
				err := cashDomain.RejectTransaction(ctx, uuid.New().String(), rejectedBy)
				So(err, ShouldNotBeNil)
			})
		})

		Convey("Cash balance calculation", func() {
			// Create categories
			incomeCode := "INC-" + time.Now().Format("20060102150405")
			incomeInput := model.CashCategoryInput{
				Code:     &incomeCode,
				Name:     func() *string { s := "Income"; return &s }(),
				Type:     func() *string { s := "income"; return &s }(),
				IsActive: func() *bool { b := true; return &b }(),
			}
			incomeCategory, err := cashDomain.CreateCategory(ctx, incomeInput)
			So(err, ShouldBeNil)

			expenseCode := "EXP-" + time.Now().Format("20060102150405")
			expenseInput := model.CashCategoryInput{
				Code:     &expenseCode,
				Name:     func() *string { s := "Expense"; return &s }(),
				Type:     func() *string { s := "expense"; return &s }(),
				IsActive: func() *bool { b := true; return &b }(),
			}
			expenseCategory, err := cashDomain.CreateCategory(ctx, expenseInput)
			So(err, ShouldBeNil)

			// Create income transactions
			for i := 0; i < 3; i++ {
				transactionDate := time.Now()
				trxType := "income"
				amount := 100000.0

				input := model.CashTransactionInput{
					TransactionDate: &transactionDate,
					Type:            &trxType,
					CategoryID:      incomeCategory.ID,
					Amount:          &amount,
				}

				_, err := cashDomain.CreateTransaction(ctx, input)
				So(err, ShouldBeNil)
			}

			// Create expense transactions
			for i := 0; i < 2; i++ {
				transactionDate := time.Now()
				trxType := "expense"
				amount := 50000.0

				input := model.CashTransactionInput{
					TransactionDate: &transactionDate,
					Type:            &trxType,
					CategoryID:      expenseCategory.ID,
					Amount:          &amount,
				}

				_, err := cashDomain.CreateTransaction(ctx, input)
				So(err, ShouldBeNil)
			}

			Convey("Get income transactions", func() {
				incomeType := "income"
				filter := model.CashTransactionFilter{
					Type: &incomeType,
				}
				transactions, err := cashDomain.ListTransactions(ctx, filter)
				So(err, ShouldBeNil)
				So(len(transactions), ShouldBeGreaterThanOrEqualTo, 3)

				totalIncome := 0.0
				for _, t := range transactions {
					totalIncome += t.Amount
				}
				So(totalIncome, ShouldBeGreaterThanOrEqualTo, 300000.0)
			})

			Convey("Get expense transactions", func() {
				expenseType := "expense"
				filter := model.CashTransactionFilter{
					Type: &expenseType,
				}
				transactions, err := cashDomain.ListTransactions(ctx, filter)
				So(err, ShouldBeNil)
				So(len(transactions), ShouldBeGreaterThanOrEqualTo, 2)

				totalExpense := 0.0
				for _, t := range transactions {
					totalExpense += t.Amount
				}
				So(totalExpense, ShouldBeGreaterThanOrEqualTo, 100000.0)
			})
		})
	})
}
