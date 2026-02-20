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
	-- ============================================
	-- INVOICES TABLE
	-- ============================================
	CREATE TABLE IF NOT EXISTS invoices (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		invoice_number VARCHAR(50) UNIQUE NOT NULL,
		customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE RESTRICT,

		-- Billing Period
		billing_period_start DATE NOT NULL,
		billing_period_end DATE NOT NULL,
		billing_month INTEGER CHECK (billing_month BETWEEN 1 AND 12),
		billing_year INTEGER,

		-- Dates
		issue_date DATE NOT NULL DEFAULT CURRENT_DATE,
		due_date DATE NOT NULL,
		payment_deadline TIMESTAMP, -- with time

		-- Amounts
		subtotal DECIMAL(12,2) NOT NULL DEFAULT 0,
		tax_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
		discount_amount DECIMAL(12,2) DEFAULT 0,
		late_fee DECIMAL(12,2) DEFAULT 0,
		total_amount DECIMAL(12,2) NOT NULL,
		paid_amount DECIMAL(12,2) DEFAULT 0,
		balance DECIMAL(12,2) GENERATED ALWAYS AS (total_amount - paid_amount) STORED,

		-- Status
		status VARCHAR(20) NOT NULL CHECK (status IN ('draft', 'sent', 'partial', 'paid', 'overdue', 'cancelled', 'refunded')) DEFAULT 'draft',
		payment_status VARCHAR(20) CHECK (payment_status IN ('unpaid', 'partial', 'paid', 'overpaid')),

		-- Payment Information
		payment_date TIMESTAMP,
		payment_method VARCHAR(20),

		-- Additional
		invoice_type VARCHAR(20) DEFAULT 'recurring' CHECK (invoice_type IN ('recurring', 'installation', 'additional', 'refund')),
		is_auto_generated BOOLEAN DEFAULT true,
		reminder_sent_count INTEGER DEFAULT 0,
		last_reminder_sent TIMESTAMP,

		notes TEXT,
		internal_notes TEXT, -- notes untuk internal, tidak tampil di PDF

		created_by UUID, -- FK to users
		updated_by UUID,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		deleted_at TIMESTAMP
	);

	-- Indexes
	CREATE INDEX IF NOT EXISTS idx_invoices_customer ON invoices(customer_id) WHERE deleted_at IS NULL;
	CREATE INDEX IF NOT EXISTS idx_invoices_status ON invoices(status) WHERE deleted_at IS NULL;
	CREATE INDEX IF NOT EXISTS idx_invoices_due_date ON invoices(due_date) WHERE deleted_at IS NULL;
	CREATE INDEX IF NOT EXISTS idx_invoices_number ON invoices(invoice_number);
	CREATE INDEX IF NOT EXISTS idx_invoices_billing_period ON invoices(billing_year, billing_month) WHERE deleted_at IS NULL;
	CREATE INDEX IF NOT EXISTS idx_invoices_payment_status ON invoices(payment_status) WHERE deleted_at IS NULL;
	CREATE INDEX IF NOT EXISTS idx_invoices_deleted ON invoices(deleted_at);

	-- Comments
	COMMENT ON TABLE invoices IS 'Tagihan/invoice pelanggan';
	COMMENT ON COLUMN invoices.balance IS 'Calculated column: total_amount - paid_amount';
	COMMENT ON COLUMN invoices.payment_status IS 'unpaid=belum bayar, partial=sebagian bayar, paid=lunas, overpaid=lebih bayar';

	-- ============================================
	-- INVOICE ITEMS TABLE
	-- ============================================
	CREATE TABLE IF NOT EXISTS invoice_items (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,

		item_type VARCHAR(20) CHECK (item_type IN ('subscription', 'installation', 'equipment', 'other')),
		description VARCHAR(255) NOT NULL,

		profile_id UUID REFERENCES bandwidth_profiles(id) ON DELETE SET NULL,

		quantity INTEGER NOT NULL DEFAULT 1,
		unit_price DECIMAL(12,2) NOT NULL,
		subtotal DECIMAL(12,2) NOT NULL,
		tax_rate DECIMAL(5,4) DEFAULT 0,
		tax_amount DECIMAL(12,2) DEFAULT 0,
		total DECIMAL(12,2) NOT NULL,

		-- Proration
		is_prorated BOOLEAN DEFAULT false,
		proration_days INTEGER,
		proration_percentage DECIMAL(5,2),

		sort_order INTEGER DEFAULT 0,

		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
	);

	-- Indexes
	CREATE INDEX IF NOT EXISTS idx_invoice_items_invoice ON invoice_items(invoice_id);
	CREATE INDEX IF NOT EXISTS idx_invoice_items_profile ON invoice_items(profile_id) WHERE profile_id IS NOT NULL;

	-- Comments
	COMMENT ON TABLE invoice_items IS 'Detail item dalam invoice';
	`)
	if err != nil {
		return err
	}
	return nil
}

func downInvoices(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	DROP TABLE IF EXISTS invoice_items CASCADE;
	DROP TABLE IF EXISTS invoices CASCADE;
	`)
	if err != nil {
		return err
	}
	return nil
}
