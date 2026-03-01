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
	-- ============================================
	-- ACTIVE SESSIONS TABLE
	-- ============================================
	CREATE TABLE IF NOT EXISTS active_sessions (
		id              UUID      PRIMARY KEY DEFAULT gen_random_uuid(),
		subscription_id UUID      NOT NULL REFERENCES subscriptions(id) ON DELETE RESTRICT,
		router_id       UUID      NOT NULL REFERENCES mikrotik_routers(id) ON DELETE RESTRICT,
		session_id      VARCHAR(100) NOT NULL, -- ID sesi dari MikroTik
		caller_id       VARCHAR(50),           -- MAC address / caller-id dari MikroTik
		ip_address      VARCHAR(15) NOT NULL,
		bytes_in        BIGINT    DEFAULT 0,
		bytes_out       BIGINT    DEFAULT 0,
		uptime_seconds  INTEGER   DEFAULT 0,
		connected_at    TIMESTAMP NOT NULL,
		last_updated    TIMESTAMP NOT NULL
	);

	CREATE UNIQUE INDEX IF NOT EXISTS uk_active_sessions_router_session ON active_sessions(router_id, session_id);
	CREATE INDEX IF NOT EXISTS idx_active_sessions_subscription ON active_sessions(subscription_id);
	CREATE INDEX IF NOT EXISTS idx_active_sessions_router       ON active_sessions(router_id);
	CREATE INDEX IF NOT EXISTS idx_active_sessions_ip           ON active_sessions(ip_address);

	COMMENT ON TABLE  active_sessions          IS 'Sesi aktif pelanggan, hasil polling dari MikroTik';
	COMMENT ON COLUMN active_sessions.session_id IS 'ID sesi unik dari RouterOS (.id pada /ppp/active)';

	-- ============================================
	-- SESSION HISTORY TABLE
	-- ============================================
	CREATE TABLE IF NOT EXISTS session_history (
		id              UUID      PRIMARY KEY DEFAULT gen_random_uuid(),
		subscription_id UUID      NOT NULL REFERENCES subscriptions(id) ON DELETE RESTRICT,
		router_id       UUID      NOT NULL REFERENCES mikrotik_routers(id) ON DELETE RESTRICT,
		session_id      VARCHAR(100),
		ip_address      VARCHAR(15),
		bytes_in        BIGINT    DEFAULT 0,
		bytes_out       BIGINT    DEFAULT 0,
		uptime_seconds  INTEGER   DEFAULT 0,
		connected_at    TIMESTAMP NOT NULL,
		disconnected_at TIMESTAMP NOT NULL,
		disconnect_cause VARCHAR(100) -- penyebab disconnect dari MikroTik
	);

	CREATE INDEX IF NOT EXISTS idx_session_history_subscription    ON session_history(subscription_id);
	CREATE INDEX IF NOT EXISTS idx_session_history_router          ON session_history(router_id);
	CREATE INDEX IF NOT EXISTS idx_session_history_disconnected_at ON session_history(disconnected_at DESC);

	COMMENT ON TABLE session_history IS 'Histori sesi pelanggan setelah disconnect — untuk analisis traffic dan billing';
	`)
	return err
}

func downActiveSessions(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	DROP TABLE IF EXISTS session_history  CASCADE;
	DROP TABLE IF EXISTS active_sessions  CASCADE;
	`)
	return err
}
