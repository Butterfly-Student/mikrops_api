//go:build integration
// +build integration

package integration_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	_ "github.com/lib/pq"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"

	postgres_outbound_adapter "mikrops/internal/adapter/outbound/postgres"
	"mikrops/internal/model"
)

func TestBillingIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:14-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		t.Fatalf("Failed to start postgres container: %v", err)
	}
	defer pgContainer.Terminate(ctx)

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to get connection string: %v", err)
	}

	db, err := gorm.Open(gormpostgres.Open(connStr), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get sql.DB: %v", err)
	}
	defer sqlDB.Close()

	// Create all required tables for billing flow
	createTablesSQL := `
		-- Tenant table
		CREATE TABLE IF NOT EXISTS tenants (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			slug VARCHAR(100) UNIQUE NOT NULL,
			email VARCHAR(255),
			phone VARCHAR(50),
			address TEXT,
			logo_url VARCHAR(500),
			max_nas INTEGER DEFAULT 3,
			subscription_plan VARCHAR(50) DEFAULT 'basic',
			subscription_expires_at TIMESTAMP,
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
		);

		-- Internet packages table
		CREATE TABLE IF NOT EXISTS internet_packages (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			type VARCHAR(50) NOT NULL,
			price BIGINT NOT NULL,
			billing_cycle VARCHAR(50) DEFAULT 'monthly',
			validity_days INTEGER DEFAULT 30,
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
		);

		-- Customer table
		CREATE TABLE IF NOT EXISTS customers (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
			full_name VARCHAR(255) NOT NULL,
			email VARCHAR(255),
			phone VARCHAR(50),
			username VARCHAR(100),
			password_hash VARCHAR(255),
			pppoe_username VARCHAR(100),
			pppoe_password VARCHAR(255),
			is_active BOOLEAN DEFAULT true,
			registered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
		);

		-- Subscription table
		CREATE TABLE IF NOT EXISTS subscriptions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
			customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
			package_id UUID NOT NULL REFERENCES internet_packages(id),
			status VARCHAR(50) DEFAULT 'active',
			start_date TIMESTAMP NOT NULL,
			end_date TIMESTAMP NOT NULL,
			auto_renew BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
		);

		-- Invoice table
		CREATE TABLE IF NOT EXISTS invoices (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
			customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
			subscription_id UUID NOT NULL REFERENCES subscriptions(id),
			invoice_number VARCHAR(100) UNIQUE NOT NULL,
			amount BIGINT NOT NULL,
			tax_amount BIGINT DEFAULT 0,
			total_amount BIGINT NOT NULL,
			status VARCHAR(50) DEFAULT 'unpaid',
			due_date TIMESTAMP NOT NULL,
			period_start TIMESTAMP NOT NULL,
			period_end TIMESTAMP NOT NULL,
			paid_at TIMESTAMP,
			notes TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
		);

		-- Payment method table
		CREATE TABLE IF NOT EXISTS payment_methods (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			type VARCHAR(50) NOT NULL,
			account_name VARCHAR(255),
			account_number VARCHAR(100),
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
		);

		-- Payment table
		CREATE TABLE IF NOT EXISTS payments (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
			invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
			payment_method_id UUID NOT NULL REFERENCES payment_methods(id),
			amount BIGINT NOT NULL,
			payment_date TIMESTAMP NOT NULL,
			proof_url VARCHAR(500),
			status VARCHAR(50) DEFAULT 'pending',
			verified_at TIMESTAMP,
			notes TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
		);

		-- Create indexes
		CREATE INDEX IF NOT EXISTS idx_internet_packages_tenant ON internet_packages(tenant_id);
		CREATE INDEX IF NOT EXISTS idx_customers_tenant ON customers(tenant_id);
		CREATE INDEX IF NOT EXISTS idx_subscriptions_tenant ON subscriptions(tenant_id);
		CREATE INDEX IF NOT EXISTS idx_subscriptions_customer ON subscriptions(customer_id);
		CREATE INDEX IF NOT EXISTS idx_invoices_tenant ON invoices(tenant_id);
		CREATE INDEX IF NOT EXISTS idx_invoices_customer ON invoices(customer_id);
		CREATE INDEX IF NOT EXISTS idx_payments_tenant ON payments(tenant_id);
		CREATE INDEX IF NOT EXISTS idx_payments_invoice ON payments(invoice_id);
	`

	_, err = sqlDB.Exec(createTablesSQL)
	if err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	Convey("End-to-End Billing Flow Integration Test", t, func() {
		tenantAdapter := postgres_outbound_adapter.NewTenantAdapter(db)
		packageAdapter := postgres_outbound_adapter.NewInternetPackageAdapter(db)
		customerAdapter := postgres_outbound_adapter.NewCustomerAdapter(db)
		subscriptionAdapter := postgres_outbound_adapter.NewSubscriptionAdapter(db)
		invoiceAdapter := postgres_outbound_adapter.NewInvoiceAdapter(db)
		paymentMethodAdapter := postgres_outbound_adapter.NewPaymentMethodAdapter(db)
		paymentAdapter := postgres_outbound_adapter.NewPaymentAdapter(db)

		now := time.Now()
		testSuffix := fmt.Sprintf("%d", now.UnixNano())

		Convey("Complete billing cycle: tenant → customer → subscription → invoice → payment", func() {
			// Step 1: Create Tenant
			tenantInput := model.TenantInput{
				Name:      "ISP Test Company",
				Slug:      "isp-test-" + testSuffix,
				Email:     "admin@isptest.com",
				MaxNas:    3,
				IsActive:  true,
				CreatedAt: now,
				UpdatedAt: now,
			}
			tenant, err := tenantAdapter.Create(tenantInput)
			So(err, ShouldBeNil)
			So(tenant.ID, ShouldNotBeEmpty)

			// Step 2: Create Internet Package
			packageInput := model.InternetPackageInput{
				TenantID:     tenant.ID,
				Name:         "Paket 10Mbps",
				Type:         "pppoe",
				Price:        300000,
				BillingCycle: "monthly",
				ValidityDays: 30,
				IsActive:     true,
				CreatedAt:    now,
				UpdatedAt:    now,
			}
			pkg, err := packageAdapter.Create(packageInput)
			So(err, ShouldBeNil)
			So(pkg.ID, ShouldNotBeEmpty)

			// Step 3: Create Customer
			customerInput := model.CustomerInput{
				TenantID:      tenant.ID,
				FullName:      "John Doe",
				Email:         "john@example.com",
				Phone:         "+6281234567890",
				Username:      "johndoe",
				PasswordHash:  "$2a$10$hash",
				IsActive:      true,
				RegisteredAt:  now,
				CreatedAt:     now,
				UpdatedAt:     now,
			}
			customer, err := customerAdapter.Create(customerInput)
			So(err, ShouldBeNil)
			So(customer.ID, ShouldNotBeEmpty)

			// Step 4: Create Subscription
			subscriptionInput := model.SubscriptionInput{
				TenantID:   tenant.ID,
				CustomerID: customer.ID,
				PackageID:  pkg.ID,
				NasID:      "dummy-nas-id", // In real scenario, this would be an actual NAS ID
				Status:     model.SubscriptionStatusActive,
				StartDate:  now,
				EndDate:    now.AddDate(0, 1, 0),
				AutoRenew:  true,
				CreatedAt:  now,
				UpdatedAt:  now,
			}
			subscription, err := subscriptionAdapter.Create(subscriptionInput)
			So(err, ShouldBeNil)
			So(subscription.ID, ShouldNotBeEmpty)

			// Step 5: Generate Invoice
			invoiceInput := model.InvoiceInput{
				TenantID:       tenant.ID,
				CustomerID:     customer.ID,
				SubscriptionID: subscription.ID,
				InvoiceNumber:  fmt.Sprintf("INV-%s-001", testSuffix),
				Amount:         300000,
				TaxAmount:      30000,
				TotalAmount:    330000,
				Status:         model.InvoiceStatusUnpaid,
				DueDate:        now.AddDate(0, 0, 7),
				PeriodStart:    now,
				PeriodEnd:      now.AddDate(0, 1, 0),
				Notes:          "Monthly subscription fee",
				CreatedAt:      now,
				UpdatedAt:      now,
			}
			invoice, err := invoiceAdapter.Create(invoiceInput)
			So(err, ShouldBeNil)
			So(invoice.ID, ShouldNotBeEmpty)
			So(invoice.Status, ShouldEqual, model.InvoiceStatusUnpaid)

			// Step 6: Create Payment Method
			paymentMethodInput := model.PaymentMethodInput{
				TenantID:      tenant.ID,
				Name:          "BCA Transfer",
				Type:          "bank_transfer",
				AccountName:   "PT ISP Test",
				AccountNumber: "1234567890",
				IsActive:      true,
				CreatedAt:     now,
				UpdatedAt:     now,
			}
			paymentMethod, err := paymentMethodAdapter.Create(paymentMethodInput)
			So(err, ShouldBeNil)
			So(paymentMethod.ID, ShouldNotBeEmpty)

			// Step 7: Create Payment (pending)
			paymentInput := model.PaymentInput{
				TenantID:        tenant.ID,
				InvoiceID:       invoice.ID,
				PaymentMethodID: paymentMethod.ID,
				Amount:          330000,
				PaymentDate:     now,
				ProofURL:        "https://example.com/proof.jpg",
				Status:          model.PaymentStatusPending,
				Notes:           "Transfer from customer",
				CreatedAt:       now,
				UpdatedAt:       now,
			}
			payment, err := paymentAdapter.Create(paymentInput)
			So(err, ShouldBeNil)
			So(payment.ID, ShouldNotBeEmpty)
			So(payment.Status, ShouldEqual, model.PaymentStatusPending)

			Convey("Verify payment and update invoice status", func() {
				// Step 8: Verify Payment
				verifiedAt := time.Now()
				updatePaymentInput := paymentInput
				updatePaymentInput.Status = model.PaymentStatusVerified
				updatePaymentInput.VerifiedAt = &verifiedAt
				updatePaymentInput.UpdatedAt = time.Now()

				err := paymentAdapter.Update(payment.ID, updatePaymentInput)
				So(err, ShouldBeNil)

				// Verify payment status updated
				verifiedPayment, err := paymentAdapter.FindByID(payment.ID)
				So(err, ShouldBeNil)
				So(verifiedPayment.Status, ShouldEqual, model.PaymentStatusVerified)
				So(verifiedPayment.VerifiedAt, ShouldNotBeNil)

				// Step 9: Update Invoice status to paid
				paidAt := time.Now()
				updateInvoiceInput := invoiceInput
				updateInvoiceInput.Status = model.InvoiceStatusPaid
				updateInvoiceInput.PaidAt = &paidAt
				updateInvoiceInput.UpdatedAt = time.Now()

				err = invoiceAdapter.Update(invoice.ID, updateInvoiceInput)
				So(err, ShouldBeNil)

				// Verify invoice is now paid
				paidInvoice, err := invoiceAdapter.FindByID(invoice.ID)
				So(err, ShouldBeNil)
				So(paidInvoice.Status, ShouldEqual, model.InvoiceStatusPaid)
				So(paidInvoice.PaidAt, ShouldNotBeNil)
			})

			Convey("Reject payment scenario", func() {
				// Create another invoice and payment for rejection scenario
				invoice2Input := invoiceInput
				invoice2Input.InvoiceNumber = fmt.Sprintf("INV-%s-002", testSuffix)
				invoice2, err := invoiceAdapter.Create(invoice2Input)
				So(err, ShouldBeNil)

				payment2Input := paymentInput
				payment2Input.InvoiceID = invoice2.ID
				payment2, err := paymentAdapter.Create(payment2Input)
				So(err, ShouldBeNil)
				So(payment2.Status, ShouldEqual, model.PaymentStatusPending)

				// Reject the payment
				payment2Input.Status = model.PaymentStatusRejected
				payment2Input.Notes = "Invalid payment proof"
				payment2Input.UpdatedAt = time.Now()

				err = paymentAdapter.Update(payment2.ID, payment2Input)
				So(err, ShouldBeNil)

				// Verify payment is rejected
				rejectedPayment, err := paymentAdapter.FindByID(payment2.ID)
				So(err, ShouldBeNil)
				So(rejectedPayment.Status, ShouldEqual, model.PaymentStatusRejected)

				// Verify invoice is still unpaid
				invoice2AfterReject, err := invoiceAdapter.FindByID(invoice2.ID)
				So(err, ShouldBeNil)
				So(invoice2AfterReject.Status, ShouldEqual, model.InvoiceStatusUnpaid)
			})

			Convey("Query invoices and payments by customer", func() {
				// Find all invoices for the customer
				invoiceFilter := model.InvoiceFilter{
					CustomerIDs: []string{customer.ID},
				}
				customerInvoices, err := invoiceAdapter.FindByFilter(invoiceFilter)
				So(err, ShouldBeNil)
				So(len(customerInvoices), ShouldBeGreaterThanOrEqualTo, 1)

				// Find all payments for the customer (via invoices)
				paymentFilter := model.PaymentFilter{
					TenantIDs: []string{tenant.ID},
				}
				tenantPayments, err := paymentAdapter.FindByFilter(paymentFilter)
				So(err, ShouldBeNil)
				So(len(tenantPayments), ShouldBeGreaterThanOrEqualTo, 1)
			})
		})

		Convey("Bulk invoice generation scenario", func() {
			// Create tenant
			tenantInput := model.TenantInput{
				Name:      "Bulk ISP",
				Slug:      fmt.Sprintf("bulk-isp-%d", time.Now().UnixNano()),
				Email:     "admin@bulkisp.com",
				IsActive:  true,
				CreatedAt: now,
				UpdatedAt: now,
			}
			tenant, err := tenantAdapter.Create(tenantInput)
			So(err, ShouldBeNil)

			// Create package
			pkg, _ := packageAdapter.Create(model.InternetPackageInput{
				TenantID:     tenant.ID,
				Name:         "Standard Package",
				Type:         "pppoe",
				Price:        200000,
				BillingCycle: "monthly",
				IsActive:     true,
				CreatedAt:    now,
				UpdatedAt:    now,
			})

			// Create multiple customers and subscriptions
			var invoices []model.InvoiceInput
			for i := 0; i < 3; i++ {
				customer, _ := customerAdapter.Create(model.CustomerInput{
					TenantID:     tenant.ID,
					FullName:     fmt.Sprintf("Customer %d", i+1),
					Email:        fmt.Sprintf("customer%d@test.com", i+1),
					IsActive:     true,
					RegisteredAt: now,
					CreatedAt:    now,
					UpdatedAt:    now,
				})

				subscription, _ := subscriptionAdapter.Create(model.SubscriptionInput{
					TenantID:   tenant.ID,
					CustomerID: customer.ID,
					PackageID:  pkg.ID,
					NasID:      "dummy-nas",
					Status:     model.SubscriptionStatusActive,
					StartDate:  now,
					EndDate:    now.AddDate(0, 1, 0),
					CreatedAt:  now,
					UpdatedAt:  now,
				})

				invoices = append(invoices, model.InvoiceInput{
					TenantID:       tenant.ID,
					CustomerID:     customer.ID,
					SubscriptionID: subscription.ID,
					InvoiceNumber:  fmt.Sprintf("INV-BULK-%s-%03d", testSuffix, i+1),
					Amount:         200000,
					TaxAmount:      20000,
					TotalAmount:    220000,
					Status:         model.InvoiceStatusUnpaid,
					DueDate:        now.AddDate(0, 0, 7),
					PeriodStart:    now,
					PeriodEnd:      now.AddDate(0, 1, 0),
					CreatedAt:      now,
					UpdatedAt:      now,
				})
			}

			// Bulk create invoices
			err = invoiceAdapter.BulkCreate(invoices)
			So(err, ShouldBeNil)

			// Verify all invoices were created
			filter := model.InvoiceFilter{TenantIDs: []string{tenant.ID}}
			createdInvoices, err := invoiceAdapter.FindByFilter(filter)
			So(err, ShouldBeNil)
			So(len(createdInvoices), ShouldEqual, 3)
		})
	})
}
