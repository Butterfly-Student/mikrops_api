package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upMikrotikSyncQueue, downMikrotikSyncQueue)
}

func upMikrotikSyncQueue(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	CREATE TABLE IF NOT EXISTS mikrotik_sync_queue (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		router_id UUID NOT NULL REFERENCES routers(id) ON DELETE RESTRICT,
		subscription_id UUID REFERENCES subscriptions(id) ON DELETE SET NULL,
		operation VARCHAR(20) NOT NULL CHECK (operation IN ('add', 'update', 'delete', 'enable', 'disable')),
		resource_type VARCHAR(20) NOT NULL CHECK (resource_type IN ('ppp_secret', 'hotspot_user', 'ip_binding', 'queue', 'address')),
		payload JSONB NOT NULL,
		priority SMALLINT DEFAULT 5,
		status VARCHAR(20) CHECK (status IN ('pending', 'processing', 'done', 'failed')) DEFAULT 'pending',
		attempts SMALLINT DEFAULT 0,
		max_attempts SMALLINT DEFAULT 3,
		error_message TEXT,
		scheduled_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		processed_at TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_mikrotik_sync_queue_router ON mikrotik_sync_queue(router_id, status);
	CREATE INDEX IF NOT EXISTS idx_mikrotik_sync_queue_status ON mikrotik_sync_queue(status, scheduled_at);
	CREATE INDEX IF NOT EXISTS idx_mikrotik_sync_queue_priority ON mikrotik_sync_queue(priority, status);
	CREATE INDEX IF NOT EXISTS idx_mikrotik_sync_queue_subscription ON mikrotik_sync_queue(subscription_id);

	COMMENT ON TABLE mikrotik_sync_queue IS 'Antrian operasi MikroTik, diproses background job';
	`)
	return err
}

func downMikrotikSyncQueue(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS mikrotik_sync_queue CASCADE;`)
	return err
}
