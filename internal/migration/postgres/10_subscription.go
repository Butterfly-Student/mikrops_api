package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upSubscription, downSubscription)
}

func upSubscription(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`CREATE TABLE IF NOT EXISTS subscriptions (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		tenant_id UUID NOT NULL,
		customer_id UUID NOT NULL,
		package_id UUID NOT NULL,
		nas_id UUID NOT NULL,
		status VARCHAR(50) DEFAULT 'active',
		start_date TIMESTAMP NOT NULL,
		end_date TIMESTAMP NOT NULL,
		auto_renew BOOLEAN DEFAULT true,
		vacation_start TIMESTAMP,
		vacation_end TIMESTAMP,
		mikrotik_queue_name VARCHAR(255),
		mikrotik_secret_name VARCHAR(255),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
		FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE CASCADE,
		FOREIGN KEY (package_id) REFERENCES internet_packages(id) ON DELETE RESTRICT,
		FOREIGN KEY (nas_id) REFERENCES nas(id) ON DELETE RESTRICT
	);
	
	CREATE INDEX IF NOT EXISTS idx_subscriptions_tenant_id ON subscriptions(tenant_id);
	CREATE INDEX IF NOT EXISTS idx_subscriptions_customer_id ON subscriptions(customer_id);
	CREATE INDEX IF NOT EXISTS idx_subscriptions_package_id ON subscriptions(package_id);
	CREATE INDEX IF NOT EXISTS idx_subscriptions_status ON subscriptions(status);
	CREATE INDEX IF NOT EXISTS idx_subscriptions_end_date ON subscriptions(end_date);`)
	if err != nil {
		return err
	}
	return nil
}

func downSubscription(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS subscriptions CASCADE;`)
	if err != nil {
		return err
	}
	return nil
}
