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
	CREATE TABLE IF NOT EXISTS sequence_counters (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		name VARCHAR(50) UNIQUE NOT NULL,
		prefix VARCHAR(10),
		padding INTEGER DEFAULT 5,
		last_number INTEGER DEFAULT 0,
		reset_monthly BOOLEAN DEFAULT false,
		reset_yearly BOOLEAN DEFAULT false,
		last_reset DATE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_sequence_counters_name ON sequence_counters(name);

	COMMENT ON TABLE sequence_counters IS 'Counter nomor urut otomatis, mencegah race condition';
	COMMENT ON COLUMN sequence_counters.reset_monthly IS 'reset ke 0 tiap awal bulan?';
	COMMENT ON COLUMN sequence_counters.reset_yearly IS 'reset ke 0 tiap awal tahun?';
	`)
	return err
}

func downSequenceCounters(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS sequence_counters CASCADE;`)
	return err
}
