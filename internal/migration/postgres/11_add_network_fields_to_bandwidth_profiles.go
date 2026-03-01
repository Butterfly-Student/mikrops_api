package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upAddNetworkFieldsToBandwidthProfiles, downAddNetworkFieldsToBandwidthProfiles)
}

// upAddNetworkFieldsToBandwidthProfiles is a no-op — network fields were merged into 6_bandwidth_profiles.go.
func upAddNetworkFieldsToBandwidthProfiles(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	-- No-op: local_address, remote_address, parent_queue, dns_server
	-- already included in bandwidth_profiles table (migration 6).
	SELECT 1;
	`)
	return err
}

func downAddNetworkFieldsToBandwidthProfiles(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`SELECT 1; -- no-op`)
	return err
}
