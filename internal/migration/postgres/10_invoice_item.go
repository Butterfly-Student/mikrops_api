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
	_, err := tx.Exec(`CREATE TABLE invoice_items (
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
		is_prorated BOOLEAN DEFAULT false,
		proration_days INTEGER,
		proration_percentage DECIMAL(5,2),
		sort_order INTEGER DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
	);`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE INDEX idx_invoice_items_invoice ON invoice_items(invoice_id);`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE INDEX idx_invoice_items_profile ON invoice_items(profile_id) WHERE profile_id IS NOT NULL;`)

	return nil
}

func downInvoiceItem(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS invoice_items CASCADE;`)
	return err
}
