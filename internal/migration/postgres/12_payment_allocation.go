package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upPaymentAllocation, downPaymentAllocation)
}

func upPaymentAllocation(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	CREATE TABLE IF NOT EXISTS payment_allocations (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		payment_id UUID NOT NULL REFERENCES payments(id) ON DELETE CASCADE,
		invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
		allocated_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		UNIQUE(payment_id, invoice_id)
	);

	CREATE INDEX IF NOT EXISTS idx_payment_allocations_payment ON payment_allocations(payment_id);
	CREATE INDEX IF NOT EXISTS idx_payment_allocations_invoice ON payment_allocations(invoice_id);
	`)
	if err != nil {
		return err
	}
	return nil
}

func downPaymentAllocation(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	DROP TABLE IF EXISTS payment_allocations CASCADE;
	`)
	if err != nil {
		return err
	}
	return nil
}
