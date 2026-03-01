package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upInvoices, downInvoices)
}

func upInvoices(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	CREATE TABLE IF NOT EXISTS invoices (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		invoice_no VARCHAR(30) UNIQUE NOT NULL,
		subscription_id UUID NOT NULL REFERENCES subscriptions(id) ON DELETE RESTRICT,
		customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE RESTRICT,
		amount NUMERIC(12,2) NOT NULL,
		tax NUMERIC(12,2) DEFAULT 0,
		discount NUMERIC(12,2) DEFAULT 0,
		total NUMERIC(12,2) NOT NULL,
		period_start DATE NOT NULL,
		period_end DATE NOT NULL,
		due_date DATE NOT NULL,
		status VARCHAR(20) CHECK (status IN ('unpaid', 'paid', 'overdue', 'cancelled', 'partial')) DEFAULT 'unpaid',
		notes TEXT,
		created_by UUID REFERENCES admin_users(id),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_invoices_no ON invoices(invoice_no);
	CREATE INDEX IF NOT EXISTS idx_invoices_customer ON invoices(customer_id);
	CREATE INDEX IF NOT EXISTS idx_invoices_subscription ON invoices(subscription_id);
	CREATE INDEX IF NOT EXISTS idx_invoices_status ON invoices(status);
	CREATE INDEX IF NOT EXISTS idx_invoices_due_date ON invoices(due_date, status);

	COMMENT ON TABLE invoices IS 'Invoice tagihan pelanggan';
	`)
	return err
}

func downInvoices(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS invoices CASCADE;`)
	return err
}
