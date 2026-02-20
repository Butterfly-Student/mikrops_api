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
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		profile_code VARCHAR(50) UNIQUE NOT NULL,
		name VARCHAR(100) NOT NULL,
		description TEXT,
		category VARCHAR(20) NOT NULL CHECK (category IN ('residential', 'business', 'corporate', 'promo')),
		ppp_profile_name VARCHAR(100) NOT NULL, -- nama profile di Mikrotik

		-- Speed Configuration (dalam kbps)
		download_speed BIGINT NOT NULL,
		upload_speed BIGINT NOT NULL,

		-- Burst Configuration
		burst_download BIGINT,
		burst_upload BIGINT,
		burst_threshold INTEGER DEFAULT 80, -- percentage
		burst_time INTEGER DEFAULT 8, -- seconds

		-- Queue Configuration
		priority INTEGER DEFAULT 8 CHECK (priority BETWEEN 1 AND 8),
		queue_type VARCHAR(20) DEFAULT 'default',
		shared_users INTEGER DEFAULT 1,
		queue_name VARCHAR(20), -- for ip static

		-- Pricing
		price_monthly DECIMAL(12,2) NOT NULL,
		price_installation DECIMAL(12,2) DEFAULT 0,
		tax_rate DECIMAL(5,4) DEFAULT 0.11, -- 11% PPN

		-- Status
		is_active BOOLEAN DEFAULT true,
		is_visible BOOLEAN DEFAULT true, -- tampil di list untuk pelanggan baru
		sort_order INTEGER DEFAULT 0,

		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		deleted_at TIMESTAMP
	);

	-- Indexes
	CREATE INDEX IF NOT EXISTS idx_bandwidth_profiles_category ON bandwidth_profiles(category);
	CREATE INDEX IF NOT EXISTS idx_bandwidth_profiles_is_active ON bandwidth_profiles(is_active);
	CREATE INDEX IF NOT EXISTS idx_bandwidth_profiles_is_visible ON bandwidth_profiles(is_visible);
	CREATE INDEX IF NOT EXISTS idx_bandwidth_profiles_code ON bandwidth_profiles(profile_code);
	CREATE INDEX IF NOT EXISTS idx_bandwidth_profiles_deleted ON bandwidth_profiles(deleted_at);

	-- Comments
	COMMENT ON TABLE bandwidth_profiles IS 'Package/paket bandwidth untuk pelanggan';
	COMMENT ON COLUMN bandwidth_profiles.burst_threshold IS 'Persentase threshold untuk mulai burst';
	COMMENT ON COLUMN bandwidth_profiles.priority IS '1=highest priority, 8=lowest priority';
	COMMENT ON COLUMN bandwidth_profiles.download_speed IS 'Download speed in kbps';
	COMMENT ON COLUMN bandwidth_profiles.upload_speed IS 'Upload speed in kbps';
	`)
	if err != nil {
		return err
	}
	return nil
}

func downBandwidthProfiles(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS bandwidth_profiles CASCADE;`)
	if err != nil {
		return err
	}
	return nil
}
