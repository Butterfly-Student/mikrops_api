package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upActiveSessions, downActiveSessions)
}

func upActiveSessions(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	CREATE TABLE IF NOT EXISTS active_sessions (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		subscription_id UUID NOT NULL REFERENCES subscriptions(id) ON DELETE RESTRICT,
		router_id UUID NOT NULL REFERENCES routers(id) ON DELETE RESTRICT,
		session_id VARCHAR(100) NOT NULL,
		caller_id VARCHAR(50),
		ip_address VARCHAR(15) NOT NULL,
		bytes_in BIGINT DEFAULT 0,
		bytes_out BIGINT DEFAULT 0,
		uptime_seconds INTEGER DEFAULT 0,
		connected_at TIMESTAMP NOT NULL,
		last_updated TIMESTAMP NOT NULL
	);

	CREATE UNIQUE INDEX IF NOT EXISTS uk_active_sessions_router_session ON active_sessions(router_id, session_id);
	CREATE INDEX IF NOT EXISTS idx_active_sessions_subscription ON active_sessions(subscription_id);
	CREATE INDEX IF NOT EXISTS idx_active_sessions_router ON active_sessions(router_id);

	COMMENT ON TABLE active_sessions IS 'Sesi aktif pelanggan, hasil polling dari MikroTik';
	`)
	return err
}

func downActiveSessions(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS active_sessions CASCADE;`)
	return err
}
