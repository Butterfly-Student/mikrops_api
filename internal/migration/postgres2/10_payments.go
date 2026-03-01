package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upPayments, downPayments)
}

func upPayments(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	CREATE TABLE IF NOT EXISTS payments (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE RESTRICT,
		payment_no VARCHAR(30) UNIQUE,
		payment_method VARCHAR(20) NOT NULL CHECK (payment_method IN ('cash', 'transfer', 'qris', 'gateway', 'other')),
		gateway_name VARCHAR(50),
		gateway_trx_id VARCHAR(150),
		gateway_response JSONB,
		amount_paid NUMERIC(12,2) NOT NULL,
		paid_at TIMESTAMP NOT NULL,
		verified_by UUID REFERENCES admin_users(id),
		verified_at TIMESTAMP,
		notes TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_payments_no ON payments(payment_no);
	CREATE INDEX IF NOT EXISTS idx_payments_invoice ON payments(invoice_id);
	CREATE INDEX IF NOT EXISTS idx_payments_gateway ON payments(gateway_name, gateway_trx_id);
	CREATE INDEX IF NOT EXISTS idx_payments_paid_at ON payments(paid_at);

	COMMENT ON TABLE payments IS 'Pembayaran invoice pelanggan';
	`)
	return err
}

func downPayments(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS payments CASCADE;`)
	return err
}
