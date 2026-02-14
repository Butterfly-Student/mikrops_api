package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upMikrotik, downMikrotik)
}

func upMikrotik(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`CREATE TABLE IF NOT EXISTS mikrotik_routers (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		name VARCHAR(100) NOT NULL,
		address VARCHAR(100) NOT NULL,
		api_port INTEGER DEFAULT 8728,
		rest_port INTEGER DEFAULT 80,
		username VARCHAR(100) NOT NULL,
		password VARCHAR(255) NOT NULL,
		password_encrypted TEXT,
		use_ssl BOOLEAN DEFAULT false,
		router_os_version VARCHAR(50),
		identity VARCHAR(255),
		is_active BOOLEAN DEFAULT true,
		last_seen_at TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		deleted_at TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_mikrotik_routers_name ON mikrotik_routers(name);
	CREATE INDEX IF NOT EXISTS idx_mikrotik_routers_address ON mikrotik_routers(address);
	CREATE INDEX IF NOT EXISTS idx_mikrotik_routers_is_active ON mikrotik_routers(is_active);`)
	if err != nil {
		return err
	}
	return nil
}

func downMikrotik(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS mikrotik_routers CASCADE;`)
	if err != nil {
		return err
	}
	return nil
}
