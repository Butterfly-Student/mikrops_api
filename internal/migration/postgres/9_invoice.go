package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upInvoice, downInvoice)
}

func upInvoice(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	-- Create ENUM for invoice status
	DO $$ BEGIN
		CREATE TYPE invoice_status AS ENUM ('draft', 'sent', 'partial', 'paid', 'overdue', 'cancelled', 'refunded');
	EXCEPTION
		WHEN duplicate_object THEN null;
	END $$;

	-- Create ENUM for payment status
	DO $$ BEGIN
		CREATE TYPE payment_status AS ENUM ('unpaid', 'partial', 'paid', 'overpaid');
	EXCEPTION
		WHEN duplicate_object THEN null;
	END $$;

	-- Create ENUM for invoice type
	DO $$ BEGIN
		CREATE TYPE invoice_type AS ENUM ('recurring', 'installation', 'additional', 'refund');
	EXCEPTION
		WHEN duplicate_object THEN null;
	END $$;

	CREATE TABLE IF NOT EXISTS invoices (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		invoice_number VARCHAR(50) UNIQUE NOT NULL,
		customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
		billing_period_start DATE NOT NULL,
		billing_period_end DATE NOT NULL,
		billing_month INTEGER NOT NULL CHECK (billing_month >= 1 AND billing_month <= 12),
		billing_year INTEGER NOT NULL CHECK (billing_year >= 2000 AND billing_year <= 2100),
		issue_date DATE NOT NULL,
		due_date DATE NOT NULL,
		payment_deadline TIMESTAMP,
		subtotal DECIMAL(12,2) NOT NULL DEFAULT 0,
		tax_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
		discount_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
		late_fee DECIMAL(12,2) NOT NULL DEFAULT 0,
		total_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
		paid_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
		status invoice_status NOT NULL DEFAULT 'draft',
		payment_status payment_status NOT NULL DEFAULT 'unpaid',
		payment_date TIMESTAMP,
		payment_method VARCHAR(50),
		invoice_type invoice_type NOT NULL DEFAULT 'recurring',
		is_auto_generated BOOLEAN DEFAULT false,
		reminder_sent_count INTEGER DEFAULT 0,
		last_reminder_sent TIMESTAMP,
		notes TEXT,
		internal_notes TEXT,
		created_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
		updated_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		deleted_at TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_invoices_customer ON invoices(customer_id);
	CREATE INDEX IF NOT EXISTS idx_invoices_status ON invoices(status);
	CREATE INDEX IF NOT EXISTS idx_invoices_payment_status ON invoices(payment_status);
	CREATE INDEX IF NOT EXISTS idx_invoices_due_date ON invoices(due_date);
	CREATE INDEX IF NOT EXISTS idx_invoices_number ON invoices(invoice_number);
	CREATE INDEX IF NOT EXISTS idx_invoices_billing_period ON invoices(billing_year, billing_month);
	CREATE INDEX IF NOT EXISTS idx_invoices_issue_date ON invoices(issue_date);
	CREATE INDEX IF NOT EXISTS idx_invoices_invoice_type ON invoices(invoice_type);
	CREATE INDEX IF NOT EXISTS idx_invoices_created_by ON invoices(created_by);
	CREATE INDEX IF NOT EXISTS idx_invoices_updated_by ON invoices(updated_by);
	CREATE INDEX IF NOT EXISTS idx_invoices_deleted_at ON invoices(deleted_at);
	`)
	if err != nil {
		return err
	}
	return nil
}

func downInvoice(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	DROP TABLE IF EXISTS invoices CASCADE;
	DROP TYPE IF EXISTS invoice_status CASCADE;
	DROP TYPE IF EXISTS payment_status CASCADE;
	DROP TYPE IF EXISTS invoice_type CASCADE;
	`)
	if err != nil {
		return err
	}
	return nil
}
