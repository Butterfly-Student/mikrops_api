package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upIpAssignments, downIpAssignments)
}

func upIpAssignments(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	CREATE TABLE IF NOT EXISTS ip_assignments (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		subscription_id UUID NOT NULL REFERENCES subscriptions(id) ON DELETE RESTRICT,
		ip_pool_id UUID NOT NULL REFERENCES ip_pools(id) ON DELETE RESTRICT,
		ip_address VARCHAR(18) NOT NULL,
		mac_address VARCHAR(17),
		assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		released_at TIMESTAMP,
		is_active BOOLEAN DEFAULT true
	);

	CREATE INDEX IF NOT EXISTS idx_ip_assignments_subscription ON ip_assignments(subscription_id);
	CREATE INDEX IF NOT EXISTS idx_ip_assignments_pool ON ip_assignments(ip_pool_id);
	CREATE UNIQUE INDEX IF NOT EXISTS uk_ip_assignments_active_ip ON ip_assignments(ip_address) WHERE is_active = true;
	CREATE INDEX IF NOT EXISTS idx_ip_assignments_is_active ON ip_assignments(is_active);

	COMMENT ON TABLE ip_assignments IS 'Alokasi IP address ke pelanggan';
	`)
	return err
}

func downIpAssignments(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS ip_assignments CASCADE;`)
	return err
}
