package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upSubscriptions, downSubscriptions)
}

func upSubscriptions(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	CREATE TABLE IF NOT EXISTS subscriptions (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE RESTRICT,
		plan_id UUID NOT NULL REFERENCES service_plans(id) ON DELETE RESTRICT,
		router_id UUID NOT NULL REFERENCES routers(id) ON DELETE RESTRICT,
		service_type VARCHAR(20) NOT NULL CHECK (service_type IN ('pppoe', 'hotspot', 'static_ip', 'vpn')),
		username VARCHAR(100) UNIQUE NOT NULL,
		password VARCHAR(100) NOT NULL,
		static_ip VARCHAR(18),
		gateway VARCHAR(15),
		vpn_type VARCHAR(20) CHECK (vpn_type IN ('pptp', 'l2tp', 'sstp', 'ovpn')),
		vpn_ip_pool VARCHAR(18),
		status VARCHAR(20) CHECK (status IN ('active', 'suspended', 'expired', 'pending', 'terminated')) DEFAULT 'pending',
		activated_at TIMESTAMP,
		expired_at TIMESTAMP,
		suspend_reason TEXT,
		terminated_at TIMESTAMP,
		mt_synced BOOLEAN DEFAULT false,
		mt_last_sync TIMESTAMP,
		mt_error TEXT,
		created_by UUID REFERENCES admin_users(id),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_subscriptions_customer ON subscriptions(customer_id);
	CREATE INDEX IF NOT EXISTS idx_subscriptions_plan ON subscriptions(plan_id);
	CREATE INDEX IF NOT EXISTS idx_subscriptions_router ON subscriptions(router_id);
	CREATE INDEX IF NOT EXISTS idx_subscriptions_username ON subscriptions(username);
	CREATE INDEX IF NOT EXISTS idx_subscriptions_status ON subscriptions(status);
	CREATE INDEX IF NOT EXISTS idx_subscriptions_expired ON subscriptions(expired_at, status);
	CREATE INDEX IF NOT EXISTS idx_subscriptions_mt_sync ON subscriptions(mt_synced);

	COMMENT ON TABLE subscriptions IS 'Layanan aktif pelanggan, jembatan ke MikroTik';
	`)
	return err
}

func downSubscriptions(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS subscriptions CASCADE;`)
	return err
}
