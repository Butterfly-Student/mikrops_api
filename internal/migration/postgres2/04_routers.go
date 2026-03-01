package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upRouters, downRouters)
}

func upRouters(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	CREATE TABLE IF NOT EXISTS routers (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		name VARCHAR(100) NOT NULL,
		ip_address VARCHAR(15) NOT NULL,
		api_port INTEGER DEFAULT 8728,
		api_username VARCHAR(100) NOT NULL,
		api_password TEXT NOT NULL,
		area VARCHAR(100),
		is_master BOOLEAN DEFAULT false,
		status VARCHAR(20) CHECK (status IN ('online', 'offline', 'unknown')) DEFAULT 'unknown',
		last_ping TIMESTAMP,
		notes TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_routers_name ON routers(name);
	CREATE INDEX IF NOT EXISTS idx_routers_ip ON routers(ip_address);
	CREATE INDEX IF NOT EXISTS idx_routers_status ON routers(status);
	CREATE INDEX IF NOT EXISTS idx_routers_is_master ON routers(is_master);

	COMMENT ON TABLE routers IS 'Daftar router MikroTik yang dikelola sistem';
	COMMENT ON COLUMN routers.api_password IS 'WAJIB dienkripsi AES-256 oleh aplikasi';
	`)
	return err
}

func downRouters(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS routers CASCADE;`)
	return err
}
