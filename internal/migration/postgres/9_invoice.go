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
	_, err := tx.Exec(`CREATE TABLE invoices (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		invoice_number VARCHAR(50) UNIQUE NOT NULL,
		customer_id UUID REFERENCES customers(id) ON DELETE RESTRICT,
		billing_period_start DATE NOT NULL,
		billing_period_end DATE NOT NULL,
		billing_month INTEGER CHECK (billing_month BETWEEN 1 AND 12),
		billing_year INTEGER,
		issue_date DATE NOT NULL DEFAULT CURRENT_DATE,
		due_date DATE NOT NULL,
		payment_deadline TIMESTAMP,
		subtotal DECIMAL(12,2) NOT NULL DEFAULT 0,
		tax_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
		discount_amount DECIMAL(12,2) DEFAULT 0,
		late_fee DECIMAL(12,2) DEFAULT 0,
		total_amount DECIMAL(12,2) NOT NULL,
		paid_amount DECIMAL(12,2) DEFAULT 0,
		balance DECIMAL(12,2) GENERATED ALWAYS AS (total_amount - paid_amount) STORED,
		status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'sent', 'partial', 'paid', 'overdue', 'cancelled', 'refunded')),
		payment_status VARCHAR(20) CHECK (payment_status IN ('unpaid', 'partial', 'paid', 'overpaid')),
		payment_date TIMESTAMP,
		payment_method VARCHAR(20),
		invoice_type VARCHAR(20) DEFAULT 'recurring' CHECK (invoice_type IN ('recurring', 'installation', 'additional', 'refund')),
		is_auto_generated BOOLEAN DEFAULT true,
		reminder_sent_count INTEGER DEFAULT 0,
		last_reminder_sent TIMESTAMP,
		notes TEXT,
		internal_notes TEXT,
		created_by UUID REFERENCES users(id),
		updated_by UUID REFERENCES users(id),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		deleted_at TIMESTAMP
	);`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE INDEX idx_invoices_customer ON invoices(customer_id) WHERE deleted_at IS NULL;`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE INDEX idx_invoices_status ON invoices(status) WHERE deleted_at IS NULL;`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE INDEX idx_invoices_due_date ON invoices(due_date) WHERE deleted_at IS NULL;`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE INDEX idx_invoices_number ON invoices(invoice_number);`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE INDEX idx_invoices_billing_period ON invoices(billing_year, billing_month) WHERE deleted_at IS NULL;`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE INDEX idx_invoices_payment_status ON invoices(payment_status) WHERE deleted_at IS NULL;`)

	return nil
}

func downInvoice(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS invoices CASCADE;`)
	return err
}
