package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upPppoeAccount, downPppoeAccount)
}

func upPppoeAccount(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`CREATE TABLE IF NOT EXISTS pppoe_accounts (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		tenant_id UUID NOT NULL,
		customer_id UUID NOT NULL,
		subscription_id UUID,
		nas_id UUID NOT NULL,
		package_id UUID NOT NULL,
		username VARCHAR(100) NOT NULL,
		password_encrypted TEXT NOT NULL,
		profile_name VARCHAR(100),
		original_profile VARCHAR(100),
		status VARCHAR(50) DEFAULT 'active',
		sync_status VARCHAR(50) DEFAULT 'pending',
		last_sync_at TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
		FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE CASCADE,
		FOREIGN KEY (subscription_id) REFERENCES subscriptions(id) ON DELETE SET NULL,
		FOREIGN KEY (nas_id) REFERENCES nas(id) ON DELETE RESTRICT,
		FOREIGN KEY (package_id) REFERENCES internet_packages(id) ON DELETE RESTRICT
	);

	CREATE UNIQUE INDEX IF NOT EXISTS idx_pppoe_accounts_tenant_username ON pppoe_accounts(tenant_id, username);
	CREATE INDEX IF NOT EXISTS idx_pppoe_accounts_tenant_id ON pppoe_accounts(tenant_id);
	CREATE INDEX IF NOT EXISTS idx_pppoe_accounts_customer_id ON pppoe_accounts(customer_id);
	CREATE INDEX IF NOT EXISTS idx_pppoe_accounts_nas_id ON pppoe_accounts(nas_id);
	CREATE INDEX IF NOT EXISTS idx_pppoe_accounts_status ON pppoe_accounts(status);
	CREATE INDEX IF NOT EXISTS idx_pppoe_accounts_sync_status ON pppoe_accounts(sync_status);`)
	if err != nil {
		return err
	}
	return nil
}

func downPppoeAccount(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS pppoe_accounts CASCADE;`)
	if err != nil {
		return err
	}
	return nil
}
