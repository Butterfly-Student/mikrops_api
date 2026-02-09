package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upActivityLog, downActivityLog)
}

func upActivityLog(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`CREATE TABLE IF NOT EXISTS activity_logs (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		tenant_id UUID NOT NULL,
		user_id UUID,
		user_type VARCHAR(50),
		action VARCHAR(100) NOT NULL,
		resource_type VARCHAR(100),
		resource_id VARCHAR(100),
		details JSONB,
		ip_address VARCHAR(45),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_activity_logs_tenant_action ON activity_logs(tenant_id, action);
	CREATE INDEX IF NOT EXISTS idx_activity_logs_tenant_resource ON activity_logs(tenant_id, resource_type, resource_id);
	CREATE INDEX IF NOT EXISTS idx_activity_logs_tenant_id ON activity_logs(tenant_id);
	CREATE INDEX IF NOT EXISTS idx_activity_logs_created_at ON activity_logs(created_at);`)
	if err != nil {
		return err
	}
	return nil
}

func downActivityLog(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS activity_logs CASCADE;`)
	if err != nil {
		return err
	}
	return nil
}
