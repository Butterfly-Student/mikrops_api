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
	-- ============================================
	-- IP POOLS TABLE
	-- ============================================
	CREATE TABLE IF NOT EXISTS ip_pools (
		id           UUID     PRIMARY KEY DEFAULT gen_random_uuid(),
		pool_name    VARCHAR(100) NOT NULL, -- harus sama persis dengan nama pool di RouterOS
		router_id    UUID     NOT NULL REFERENCES mikrotik_routers(id) ON DELETE RESTRICT,
		network      VARCHAR(18) NOT NULL,  -- CIDR, mis: 192.168.1.0/24
		gateway      VARCHAR(15) NOT NULL,
		dns_primary   VARCHAR(15) DEFAULT '8.8.8.8',
		dns_secondary VARCHAR(15) DEFAULT '8.8.4.4',
		service_type VARCHAR(20) NOT NULL CHECK (service_type IN ('pppoe', 'hotspot', 'static', 'vpn')),
		total_ip     INTEGER NOT NULL,
		used_ip      INTEGER DEFAULT 0,
		is_active    BOOLEAN DEFAULT true
	);

	CREATE INDEX IF NOT EXISTS idx_ip_pools_name   ON ip_pools(pool_name);
	CREATE INDEX IF NOT EXISTS idx_ip_pools_router ON ip_pools(router_id);
	CREATE INDEX IF NOT EXISTS idx_ip_pools_type   ON ip_pools(service_type);
	CREATE INDEX IF NOT EXISTS idx_ip_pools_active ON ip_pools(is_active);

	COMMENT ON TABLE  ip_pools          IS 'Pool IP address per router MikroTik';
	COMMENT ON COLUMN ip_pools.pool_name IS 'Nama pool HARUS sama persis dengan di RouterOS';
	COMMENT ON COLUMN ip_pools.used_ip   IS 'Counter di-update aplikasi, bukan real-time dari router';

	-- ============================================
	-- IP ASSIGNMENTS TABLE
	-- ============================================
	CREATE TABLE IF NOT EXISTS ip_assignments (
		id              UUID     PRIMARY KEY DEFAULT gen_random_uuid(),
		subscription_id UUID     NOT NULL REFERENCES subscriptions(id) ON DELETE RESTRICT,
		ip_pool_id      UUID     NOT NULL REFERENCES ip_pools(id) ON DELETE RESTRICT,
		ip_address      VARCHAR(18) NOT NULL,
		mac_address     VARCHAR(17),
		assigned_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		released_at     TIMESTAMP,
		is_active       BOOLEAN DEFAULT true
	);

	CREATE INDEX IF NOT EXISTS idx_ip_assignments_subscription ON ip_assignments(subscription_id);
	CREATE INDEX IF NOT EXISTS idx_ip_assignments_pool         ON ip_assignments(ip_pool_id);
	CREATE INDEX IF NOT EXISTS idx_ip_assignments_is_active    ON ip_assignments(is_active);
	CREATE UNIQUE INDEX IF NOT EXISTS uk_ip_assignments_active_ip ON ip_assignments(ip_address) WHERE is_active = true;

	COMMENT ON TABLE ip_assignments IS 'Alokasi IP address ke subscription pelanggan';
	`)
	return err
}

func downIpPools(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	DROP TABLE IF EXISTS ip_assignments CASCADE;
	DROP TABLE IF EXISTS ip_pools       CASCADE;
	`)
	return err
}
