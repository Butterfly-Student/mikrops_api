package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upNas, downNas)
}

func upNas(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`CREATE TABLE IF NOT EXISTS nas (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		tenant_id UUID NOT NULL,
		name VARCHAR(255) NOT NULL,
		host VARCHAR(255) NOT NULL,
		api_port INTEGER DEFAULT 8728,
		rest_port INTEGER DEFAULT 80,
		username VARCHAR(255) NOT NULL,
		password_encrypted TEXT,
		use_ssl BOOLEAN DEFAULT false,
		router_os_version VARCHAR(50),
		identity VARCHAR(255),
		is_active BOOLEAN DEFAULT true,
		last_seen_at TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
	);
	
	CREATE INDEX IF NOT EXISTS idx_nas_tenant_id ON nas(tenant_id);
	CREATE INDEX IF NOT EXISTS idx_nas_host ON nas(host);`)
	if err != nil {
		return err
	}
	return nil
}

func downNas(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS nas CASCADE;`)
	if err != nil {
		return err
	}
	return nil
}
