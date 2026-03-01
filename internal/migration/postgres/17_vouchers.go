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
	-- ============================================
	-- VOUCHER BATCHES TABLE
	-- ============================================
	CREATE TABLE IF NOT EXISTS voucher_batches (
		id             UUID     PRIMARY KEY DEFAULT gen_random_uuid(),
		batch_code     VARCHAR(50) UNIQUE NOT NULL,
		plan_id        UUID     NOT NULL REFERENCES bandwidth_profiles(id) ON DELETE RESTRICT,
		router_id      UUID     NOT NULL REFERENCES mikrotik_routers(id) ON DELETE RESTRICT,
		total_vouchers INTEGER  NOT NULL,
		sold_vouchers  INTEGER  DEFAULT 0,
		used_vouchers  INTEGER  DEFAULT 0,
		price_override DECIMAL(12,2), -- NULL = gunakan harga dari plan
		notes          TEXT,
		created_by     UUID     NOT NULL REFERENCES admin_users(id) ON DELETE RESTRICT,
		created_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_voucher_batches_code       ON voucher_batches(batch_code);
	CREATE INDEX IF NOT EXISTS idx_voucher_batches_plan       ON voucher_batches(plan_id);
	CREATE INDEX IF NOT EXISTS idx_voucher_batches_router     ON voucher_batches(router_id);
	CREATE INDEX IF NOT EXISTS idx_voucher_batches_created_by ON voucher_batches(created_by);

	COMMENT ON TABLE  voucher_batches               IS 'Batch pembuatan voucher hotspot';
	COMMENT ON COLUMN voucher_batches.price_override IS 'NULL = gunakan harga dari bandwidth_profiles';

	-- ============================================
	-- VOUCHERS TABLE
	-- ============================================
	CREATE TABLE IF NOT EXISTS vouchers (
		id         UUID      PRIMARY KEY DEFAULT gen_random_uuid(),
		batch_id   UUID      NOT NULL REFERENCES voucher_batches(id) ON DELETE RESTRICT,
		code       VARCHAR(30) UNIQUE NOT NULL,
		username   VARCHAR(100) NOT NULL, -- username di MikroTik hotspot
		password   VARCHAR(100) NOT NULL, -- password di MikroTik hotspot
		status     VARCHAR(20) CHECK (status IN ('available', 'sold', 'used', 'expired')) DEFAULT 'available',
		sold_to    VARCHAR(100), -- nama pembeli (tidak harus pelanggan terdaftar)
		sold_at    TIMESTAMP,
		used_by    VARCHAR(100), -- identifier saat voucher digunakan
		used_at    TIMESTAMP,
		expired_at TIMESTAMP,
		mt_synced  BOOLEAN DEFAULT false
	);

	CREATE INDEX IF NOT EXISTS idx_vouchers_batch    ON vouchers(batch_id);
	CREATE INDEX IF NOT EXISTS idx_vouchers_code     ON vouchers(code);
	CREATE INDEX IF NOT EXISTS idx_vouchers_status   ON vouchers(status);
	CREATE INDEX IF NOT EXISTS idx_vouchers_username ON vouchers(username);

	COMMENT ON TABLE vouchers IS 'Voucher hotspot individual';
	`)
	return err
}

func downVouchers(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	DROP TABLE IF EXISTS vouchers        CASCADE;
	DROP TABLE IF EXISTS voucher_batches CASCADE;
	`)
	return err
}
