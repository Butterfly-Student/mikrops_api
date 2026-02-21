package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upAddNetworkFieldsToBandwidthProfiles, downAddNetworkFieldsToBandwidthProfiles)
}

func upAddNetworkFieldsToBandwidthProfiles(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	ALTER TABLE bandwidth_profiles
	ADD COLUMN IF NOT EXISTS local_address VARCHAR(45),
	ADD COLUMN IF NOT EXISTS remote_address VARCHAR(45),
	ADD COLUMN IF NOT EXISTS parent_queue VARCHAR(100),
	ADD COLUMN IF NOT EXISTS dns_server VARCHAR(100);

	COMMENT ON COLUMN bandwidth_profiles.local_address IS 'PPP profile local-address (IP or pool name)';
	COMMENT ON COLUMN bandwidth_profiles.remote_address IS 'PPP profile remote-address (IP or pool name)';
	COMMENT ON COLUMN bandwidth_profiles.parent_queue IS 'PPP profile parent-queue for hierarchical QoS';
	COMMENT ON COLUMN bandwidth_profiles.dns_server IS 'DNS server assigned to PPP clients';
	`)
	return err
}

func downAddNetworkFieldsToBandwidthProfiles(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	ALTER TABLE bandwidth_profiles
	DROP COLUMN IF EXISTS local_address,
	DROP COLUMN IF EXISTS remote_address,
	DROP COLUMN IF EXISTS parent_queue,
	DROP COLUMN IF EXISTS dns_server;
	`)
	return err
}
