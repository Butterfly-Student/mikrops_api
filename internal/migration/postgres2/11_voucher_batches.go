package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upVoucherBatches, downVoucherBatches)
}

func upVoucherBatches(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	CREATE TABLE IF NOT EXISTS voucher_batches (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		batch_code VARCHAR(50) UNIQUE NOT NULL,
		plan_id UUID NOT NULL REFERENCES service_plans(id) ON DELETE RESTRICT,
		router_id UUID NOT NULL REFERENCES routers(id) ON DELETE RESTRICT,
		total_vouchers INTEGER NOT NULL,
		sold_vouchers INTEGER DEFAULT 0,
		used_vouchers INTEGER DEFAULT 0,
		notes TEXT,
		created_by UUID NOT NULL REFERENCES admin_users(id),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_voucher_batches_code ON voucher_batches(batch_code);
	CREATE INDEX IF NOT EXISTS idx_voucher_batches_plan ON voucher_batches(plan_id);
	CREATE INDEX IF NOT EXISTS idx_voucher_batches_router ON voucher_batches(router_id);
	CREATE INDEX IF NOT EXISTS idx_voucher_batches_created_by ON voucher_batches(created_by);

	COMMENT ON TABLE voucher_batches IS 'Batch pembuatan voucher hotspot';
	`)
	return err
}

func downVoucherBatches(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS voucher_batches CASCADE;`)
	return err
}
