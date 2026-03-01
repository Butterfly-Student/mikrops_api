package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upIpPools, downIpPools)
}

func upIpPools(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	CREATE TABLE IF NOT EXISTS ip_pools (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		pool_name VARCHAR(100) NOT NULL,
		router_id UUID NOT NULL REFERENCES routers(id) ON DELETE RESTRICT,
		network VARCHAR(18) NOT NULL,
		gateway VARCHAR(15) NOT NULL,
		dns_primary VARCHAR(15) DEFAULT '8.8.8.8',
		dns_secondary VARCHAR(15) DEFAULT '8.8.4.4',
		service_type VARCHAR(20) NOT NULL CHECK (service_type IN ('pppoe', 'hotspot', 'static', 'vpn')),
		total_ip INTEGER NOT NULL,
		used_ip INTEGER DEFAULT 0,
		is_active BOOLEAN DEFAULT true
	);

	CREATE INDEX IF NOT EXISTS idx_ip_pools_name ON ip_pools(pool_name);
	CREATE INDEX IF NOT EXISTS idx_ip_pools_router ON ip_pools(router_id);
	CREATE INDEX IF NOT EXISTS idx_ip_pools_type ON ip_pools(service_type);
	CREATE INDEX IF NOT EXISTS idx_ip_pools_active ON ip_pools(is_active);

	COMMENT ON TABLE ip_pools IS 'Pool IP address per router';
	COMMENT ON COLUMN ip_pools.pool_name IS 'nama pool harus sama persis dengan di RouterOS';
	`)
	return err
}

func downIpPools(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS ip_pools CASCADE;`)
	return err
}
