package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upSessionHistory, downSessionHistory)
}

func upSessionHistory(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	CREATE TABLE IF NOT EXISTS session_history (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		subscription_id UUID NOT NULL REFERENCES subscriptions(id) ON DELETE RESTRICT,
		router_id UUID NOT NULL REFERENCES routers(id) ON DELETE RESTRICT,
		session_id VARCHAR(100),
		ip_address VARCHAR(15),
		bytes_in BIGINT DEFAULT 0,
		bytes_out BIGINT DEFAULT 0,
		uptime_seconds INTEGER DEFAULT 0,
		connected_at TIMESTAMP NOT NULL,
		disconnected_at TIMESTAMP NOT NULL,
		disconnect_cause VARCHAR(100)
	);

	CREATE INDEX IF NOT EXISTS idx_session_history_subscription ON session_history(subscription_id);
	CREATE INDEX IF NOT EXISTS idx_session_history_router ON session_history(router_id);
	CREATE INDEX IF NOT EXISTS idx_session_history_disconnected ON session_history(disconnected_at);
	CREATE INDEX IF NOT EXISTS idx_session_history_disconnected_at ON session_history(disconnected_at DESC);

	COMMENT ON TABLE session_history IS 'Histori sesi pelanggan setelah disconnect';
	`)
	return err
}

func downSessionHistory(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS session_history CASCADE;`)
	return err
}
