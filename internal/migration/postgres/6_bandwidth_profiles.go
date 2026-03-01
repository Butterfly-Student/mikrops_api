package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upBandwidthProfiles, downBandwidthProfiles)
}

func upBandwidthProfiles(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	-- ============================================
	-- BANDWIDTH PROFILES TABLE
	-- ============================================
	CREATE TABLE IF NOT EXISTS bandwidth_profiles (
		id           UUID      PRIMARY KEY DEFAULT gen_random_uuid(),
		profile_code VARCHAR(50) UNIQUE NOT NULL,
		name         VARCHAR(100) NOT NULL,
		description  TEXT,

		-- Jenis layanan (multi-service support)
		service_type VARCHAR(20) NOT NULL DEFAULT 'pppoe'
		             CHECK (service_type IN ('pppoe', 'hotspot', 'static_ip', 'vpn')),
		category     VARCHAR(20) NOT NULL DEFAULT 'residential'
		             CHECK (category IN ('residential', 'business', 'corporate', 'promo')),

		-- Profil & Pool di MikroTik
		ppp_profile_name VARCHAR(100), -- nama profile PPP di Mikrotik (khusus pppoe/vpn)
		address_pool     VARCHAR(100), -- nama pool IP di MikroTik (hotspot/static_ip)

		-- Konfigurasi Speed (dalam kbps)
		download_speed BIGINT NOT NULL,
		upload_speed   BIGINT NOT NULL,

		-- Burst Configuration
		burst_download   BIGINT,
		burst_upload     BIGINT,
		burst_threshold  INTEGER DEFAULT 80,  -- persentase threshold mulai burst
		burst_time       INTEGER DEFAULT 8,   -- durasi burst dalam detik

		-- Queue Configuration
		priority     INTEGER DEFAULT 8 CHECK (priority BETWEEN 1 AND 8),
		queue_type   VARCHAR(20) DEFAULT 'default',
		shared_users INTEGER DEFAULT 1,
		queue_name   VARCHAR(100), -- nama queue untuk IP static

		-- Network (PPPoE profile detail)
		local_address  VARCHAR(45),  -- PPP local-address (IP atau nama pool)
		remote_address VARCHAR(45),  -- PPP remote-address (IP atau nama pool)
		parent_queue   VARCHAR(100), -- parent queue untuk hierarchical QoS
		dns_server     VARCHAR(100), -- DNS server untuk klien PPP

		-- Kuota (NULL = unlimited)
		quota_gb INTEGER,

		-- Siklus Tagihan
		billing_cycle VARCHAR(20) DEFAULT 'monthly'
		              CHECK (billing_cycle IN ('daily', 'weekly', 'monthly', 'yearly')),

		-- Harga
		price_monthly       DECIMAL(12,2) NOT NULL,
		price_installation  DECIMAL(12,2) DEFAULT 0,
		tax_rate            DECIMAL(5,4)  DEFAULT 0.11, -- PPN 11%

		-- Status & Tampilan
		is_active   BOOLEAN DEFAULT true,
		is_visible  BOOLEAN DEFAULT true, -- tampil di list untuk pelanggan baru
		sort_order  INTEGER DEFAULT 0,

		created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		deleted_at  TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_bandwidth_profiles_service_type ON bandwidth_profiles(service_type);
	CREATE INDEX IF NOT EXISTS idx_bandwidth_profiles_category     ON bandwidth_profiles(category);
	CREATE INDEX IF NOT EXISTS idx_bandwidth_profiles_is_active    ON bandwidth_profiles(is_active);
	CREATE INDEX IF NOT EXISTS idx_bandwidth_profiles_is_visible   ON bandwidth_profiles(is_visible);
	CREATE INDEX IF NOT EXISTS idx_bandwidth_profiles_code         ON bandwidth_profiles(profile_code);
	CREATE INDEX IF NOT EXISTS idx_bandwidth_profiles_deleted      ON bandwidth_profiles(deleted_at);

	COMMENT ON TABLE  bandwidth_profiles                IS 'Paket bandwidth untuk semua jenis layanan ISP';
	COMMENT ON COLUMN bandwidth_profiles.service_type    IS 'pppoe=PPPoE, hotspot=voucher/hotspot, static_ip=IP statik, vpn=VPN';
	COMMENT ON COLUMN bandwidth_profiles.quota_gb        IS 'Kuota dalam GB, NULL = unlimited';
	COMMENT ON COLUMN bandwidth_profiles.billing_cycle   IS 'Siklus tagihan: daily/weekly/monthly/yearly';
	COMMENT ON COLUMN bandwidth_profiles.burst_threshold IS 'Persentase threshold untuk mulai burst (default 80%)';
	COMMENT ON COLUMN bandwidth_profiles.priority        IS '1=prioritas tertinggi, 8=terendah';
	COMMENT ON COLUMN bandwidth_profiles.download_speed  IS 'Kecepatan download dalam kbps';
	COMMENT ON COLUMN bandwidth_profiles.upload_speed    IS 'Kecepatan upload dalam kbps';
	COMMENT ON COLUMN bandwidth_profiles.address_pool    IS 'Nama pool di MikroTik untuk hotspot/static_ip';
	`)
	return err
}

func downBandwidthProfiles(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS bandwidth_profiles CASCADE;`)
	return err
}
