package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upSubscriptions, downSubscriptions)
}

// upSubscriptions creates the subscriptions table.
// This is the bridge between customer <-> plan <-> router,
// replacing the old approach of storing PPPoE config directly in customers.
func upSubscriptions(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	-- ============================================
	-- SUBSCRIPTIONS TABLE
	-- ============================================
	CREATE TABLE IF NOT EXISTS subscriptions (
		id           UUID      PRIMARY KEY DEFAULT gen_random_uuid(),
		customer_id  UUID      NOT NULL REFERENCES customers(id) ON DELETE RESTRICT,
		plan_id      UUID      NOT NULL REFERENCES bandwidth_profiles(id) ON DELETE RESTRICT,
		router_id    UUID      NOT NULL REFERENCES mikrotik_routers(id) ON DELETE RESTRICT,

		-- Jenis layanan (mengikuti service_type dari bandwidth_profiles)
		service_type VARCHAR(20) NOT NULL
		             CHECK (service_type IN ('pppoe', 'hotspot', 'static_ip', 'vpn')),

		-- Kredensial di MikroTik
		username VARCHAR(100) UNIQUE NOT NULL, -- PPP username / hotspot username
		password VARCHAR(255) NOT NULL,        -- encrypted

		-- Network Config
		static_ip   VARCHAR(45),  -- IP statis opsional
		gateway     VARCHAR(15),  -- gateway override
		mac_address VARCHAR(17),  -- MAC address binding opsional
		vpn_type    VARCHAR(20)  CHECK (vpn_type IN ('pptp', 'l2tp', 'sstp', 'ovpn')),

		-- Status & Periode
		status       VARCHAR(20) NOT NULL
		             CHECK (status IN ('pending', 'active', 'suspended', 'isolated', 'expired', 'terminated'))
		             DEFAULT 'pending',
		activated_at TIMESTAMP,
		expired_at   TIMESTAMP,
		expiry_date  DATE,       -- tanggal jatuh tempo berlangganan

		-- Billing
		billing_cycle VARCHAR(20) DEFAULT 'monthly'
		              CHECK (billing_cycle IN ('daily', 'weekly', 'monthly', 'yearly')),
		billing_day   INTEGER DEFAULT 1 CHECK (billing_day BETWEEN 1 AND 31),

		-- Alasan suspend/terminate
		suspend_reason   TEXT,
		terminated_at    TIMESTAMP,

		-- Sinkronisasi ke MikroTik
		mt_synced    BOOLEAN   DEFAULT false,
		mt_last_sync TIMESTAMP,
		mt_error     TEXT,

		-- Profil sebelumnya (untuk restore setelah isolasi)
		previous_plan_id UUID REFERENCES bandwidth_profiles(id) ON DELETE SET NULL,

		notes      TEXT,
		created_by UUID REFERENCES admin_users(id) ON DELETE SET NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		deleted_at TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_subscriptions_customer    ON subscriptions(customer_id)   WHERE deleted_at IS NULL;
	CREATE INDEX IF NOT EXISTS idx_subscriptions_plan        ON subscriptions(plan_id)        WHERE deleted_at IS NULL;
	CREATE INDEX IF NOT EXISTS idx_subscriptions_router      ON subscriptions(router_id)      WHERE deleted_at IS NULL;
	CREATE INDEX IF NOT EXISTS idx_subscriptions_username    ON subscriptions(username)       WHERE deleted_at IS NULL;
	CREATE INDEX IF NOT EXISTS idx_subscriptions_status      ON subscriptions(status)         WHERE deleted_at IS NULL;
	CREATE INDEX IF NOT EXISTS idx_subscriptions_expiry      ON subscriptions(expiry_date)    WHERE deleted_at IS NULL;
	CREATE INDEX IF NOT EXISTS idx_subscriptions_mt_synced   ON subscriptions(mt_synced)      WHERE mt_synced = false;
	CREATE INDEX IF NOT EXISTS idx_subscriptions_service     ON subscriptions(service_type)   WHERE deleted_at IS NULL;
	CREATE INDEX IF NOT EXISTS idx_subscriptions_deleted     ON subscriptions(deleted_at);

	COMMENT ON TABLE  subscriptions              IS 'Layanan aktif pelanggan — jembatan antara customer, plan, dan router MikroTik';
	COMMENT ON COLUMN subscriptions.service_type  IS 'pppoe/hotspot/static_ip/vpn, harus sesuai dengan plan';
	COMMENT ON COLUMN subscriptions.username      IS 'PPP username di MikroTik (harus unik per router)';
	COMMENT ON COLUMN subscriptions.mt_synced     IS 'FALSE = perlu disinkronisasi ke MikroTik';
	COMMENT ON COLUMN subscriptions.mt_error      IS 'Pesan error terakhir saat sinkronisasi ke MikroTik gagal';
	COMMENT ON COLUMN subscriptions.previous_plan_id IS 'Plan sebelum isolasi, dikembalikan saat isolasi dicabut';
	`)
	return err
}

func downSubscriptions(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS subscriptions CASCADE;`)
	return err
}
