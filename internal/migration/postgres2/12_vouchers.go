package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upVouchers, downVouchers)
}

func upVouchers(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	CREATE TABLE IF NOT EXISTS vouchers (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		batch_id UUID NOT NULL REFERENCES voucher_batches(id) ON DELETE RESTRICT,
		code VARCHAR(30) UNIQUE NOT NULL,
		username VARCHAR(100) NOT NULL,
		password VARCHAR(100) NOT NULL,
		status VARCHAR(20) CHECK (status IN ('available', 'sold', 'used', 'expired')) DEFAULT 'available',
		sold_to VARCHAR(100),
		sold_at TIMESTAMP,
		used_by VARCHAR(100),
		used_at TIMESTAMP,
		expired_at TIMESTAMP,
		mt_synced BOOLEAN DEFAULT false
	);

	CREATE INDEX IF NOT EXISTS idx_vouchers_batch ON vouchers(batch_id);
	CREATE INDEX IF NOT EXISTS idx_vouchers_code ON vouchers(code);
	CREATE INDEX IF NOT EXISTS idx_vouchers_status ON vouchers(status);
	CREATE INDEX IF NOT EXISTS idx_vouchers_username ON vouchers(username);

	COMMENT ON TABLE vouchers IS 'Voucher hotspot individual';
	`)
	return err
}

func downVouchers(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS vouchers CASCADE;`)
	return err
}
