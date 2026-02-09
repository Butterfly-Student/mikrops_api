package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upMikrotikSyncLog, downMikrotikSyncLog)
}

func upMikrotikSyncLog(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`CREATE TABLE IF NOT EXISTS mikrotik_sync_logs (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		tenant_id UUID NOT NULL,
		nas_id UUID NOT NULL,
		action VARCHAR(50) NOT NULL,
		resource_type VARCHAR(50) NOT NULL,
		resource_identifier VARCHAR(255),
		status VARCHAR(50) NOT NULL,
		error_message TEXT,
		request_payload JSONB,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
		FOREIGN KEY (nas_id) REFERENCES nas(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_mikrotik_sync_logs_nas_resource ON mikrotik_sync_logs(nas_id, resource_type);
	CREATE INDEX IF NOT EXISTS idx_mikrotik_sync_logs_tenant_created ON mikrotik_sync_logs(tenant_id, created_at);
	CREATE INDEX IF NOT EXISTS idx_mikrotik_sync_logs_tenant_id ON mikrotik_sync_logs(tenant_id);
	CREATE INDEX IF NOT EXISTS idx_mikrotik_sync_logs_status ON mikrotik_sync_logs(status);`)
	if err != nil {
		return err
	}
	return nil
}

func downMikrotikSyncLog(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS mikrotik_sync_logs CASCADE;`)
	if err != nil {
		return err
	}
	return nil
}
