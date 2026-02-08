package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upInternetPackage, downInternetPackage)
}

func upInternetPackage(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`CREATE TABLE IF NOT EXISTS internet_packages (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		tenant_id UUID NOT NULL,
		name VARCHAR(255) NOT NULL,
		description TEXT,
		type VARCHAR(50) NOT NULL,
		upload_rate VARCHAR(50),
		download_rate VARCHAR(50),
		upload_burst VARCHAR(50),
		download_burst VARCHAR(50),
		price BIGINT NOT NULL,
		billing_cycle VARCHAR(50) DEFAULT 'monthly',
		validity_days INTEGER DEFAULT 30,
		is_active BOOLEAN DEFAULT true,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
	);
	
	CREATE INDEX IF NOT EXISTS idx_internet_packages_tenant_id ON internet_packages(tenant_id);
	CREATE INDEX IF NOT EXISTS idx_internet_packages_type ON internet_packages(type);`)
	if err != nil {
		return err
	}
	return nil
}

func downInternetPackage(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS internet_packages CASCADE;`)
	if err != nil {
		return err
	}
	return nil
}
