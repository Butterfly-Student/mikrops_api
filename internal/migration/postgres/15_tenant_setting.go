package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upTenantSetting, downTenantSetting)
}

func upTenantSetting(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`CREATE TABLE IF NOT EXISTS tenant_settings (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		tenant_id UUID NOT NULL UNIQUE,
		cutoff_day INT DEFAULT 1,
		grace_period INT DEFAULT 3,
		isolir_profile_name VARCHAR(100) DEFAULT 'ISOLIR_MIKROPS',
		wa_gateway_api TEXT,
		auto_approve_registration BOOLEAN DEFAULT false,
		default_ppp_password_type VARCHAR(20) DEFAULT 'random',
		default_ppp_password_length INT DEFAULT 8,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_tenant_settings_tenant_id ON tenant_settings(tenant_id);`)
	if err != nil {
		return err
	}
	return nil
}

func downTenantSetting(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS tenant_settings CASCADE;`)
	if err != nil {
		return err
	}
	return nil
}
