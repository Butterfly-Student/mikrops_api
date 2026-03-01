package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upSequenceCounters, downSequenceCounters)
}

func upSequenceCounters(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	-- ============================================
	-- SEQUENCE COUNTERS TABLE
	-- ============================================
	CREATE TABLE IF NOT EXISTS sequence_counters (
		id            UUID     PRIMARY KEY DEFAULT gen_random_uuid(),
		name          VARCHAR(50) UNIQUE NOT NULL, -- mis: 'invoice', 'payment', 'customer_code'
		prefix        VARCHAR(10),                 -- mis: 'INV', 'PAY', 'CST'
		padding       INTEGER DEFAULT 5,           -- jumlah digit angka
		last_number   INTEGER DEFAULT 0,
		reset_monthly BOOLEAN DEFAULT false,       -- reset ke 0 tiap awal bulan
		reset_yearly  BOOLEAN DEFAULT false,       -- reset ke 0 tiap awal tahun
		last_reset    DATE,
		created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		updated_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_sequence_counters_name ON sequence_counters(name);

	-- Seed data awal untuk nomor urut utama
	INSERT INTO sequence_counters (name, prefix, padding, reset_yearly) VALUES
		('invoice',       'INV', 5, true),
		('payment',       'PAY', 5, true),
		('customer_code', 'CST', 4, false);

	COMMENT ON TABLE  sequence_counters              IS 'Counter nomor urut otomatis, thread-safe dengan SELECT FOR UPDATE';
	COMMENT ON COLUMN sequence_counters.reset_monthly IS 'Reset ke 0 setiap awal bulan';
	COMMENT ON COLUMN sequence_counters.reset_yearly  IS 'Reset ke 0 setiap awal tahun';
	COMMENT ON COLUMN sequence_counters.last_number   IS 'Nomor terakhir yang dipakai';
	`)
	return err
}

func downSequenceCounters(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS sequence_counters CASCADE;`)
	return err
}
