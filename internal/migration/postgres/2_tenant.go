package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upTenant, downTenant)
}

func upTenant(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`CREATE TABLE IF NOT EXISTS tenants (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		name VARCHAR(255) NOT NULL,
		slug VARCHAR(100) UNIQUE NOT NULL,
		email VARCHAR(255),
		phone VARCHAR(50),
		address TEXT,
		logo_url VARCHAR(500),
		max_nas INTEGER DEFAULT 3,
		subscription_plan VARCHAR(50) DEFAULT 'basic',
		subscription_expires_at TIMESTAMP,
		is_active BOOLEAN DEFAULT true,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
	);
	
	CREATE INDEX IF NOT EXISTS idx_tenants_slug ON tenants(slug);
	CREATE INDEX IF NOT EXISTS idx_tenants_email ON tenants(email);`)
	if err != nil {
		return err
	}
	return nil
}

func downTenant(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS tenants CASCADE;`)
	if err != nil {
		return err
	}
	return nil
}
