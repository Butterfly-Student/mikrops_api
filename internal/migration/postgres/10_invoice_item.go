package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upInvoiceItem, downInvoiceItem)
}

func upInvoiceItem(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	-- Create ENUM for invoice item type
	DO $$ BEGIN
		CREATE TYPE invoice_item_type AS ENUM ('subscription', 'installation', 'equipment', 'other');
	EXCEPTION
		WHEN duplicate_object THEN null;
	END $$;

	CREATE TABLE IF NOT EXISTS invoice_items (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
		item_type invoice_item_type NOT NULL DEFAULT 'subscription',
		description VARCHAR(500) NOT NULL,
		profile_id UUID REFERENCES bandwidth_profiles(id) ON DELETE SET NULL,
		quantity INTEGER NOT NULL DEFAULT 1 CHECK (quantity > 0),
		unit_price DECIMAL(12,2) NOT NULL DEFAULT 0,
		subtotal DECIMAL(12,2) NOT NULL DEFAULT 0,
		tax_rate DECIMAL(5,4) NOT NULL DEFAULT 0.1100,
		tax_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
		total DECIMAL(12,2) NOT NULL DEFAULT 0,
		is_prorated BOOLEAN DEFAULT false,
		proration_days INTEGER,
		proration_percentage DECIMAL(5,2),
		sort_order INTEGER DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_invoice_items_invoice ON invoice_items(invoice_id);
	CREATE INDEX IF NOT EXISTS idx_invoice_items_profile ON invoice_items(profile_id);
	CREATE INDEX IF NOT EXISTS idx_invoice_items_type ON invoice_items(item_type);
	CREATE INDEX IF NOT EXISTS idx_invoice_items_sort ON invoice_items(sort_order);
	`)
	if err != nil {
		return err
	}
	return nil
}

func downInvoiceItem(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	DROP TABLE IF EXISTS invoice_items CASCADE;
	DROP TYPE IF EXISTS invoice_item_type CASCADE;
	`)
	if err != nil {
		return err
	}
	return nil
}
