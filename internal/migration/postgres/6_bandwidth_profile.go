package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upBandwidthProfile, downBandwidthProfile)
}

func upBandwidthProfile(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	-- Create ENUM for bandwidth profile category
	DO $$ BEGIN
		CREATE TYPE bandwidth_profile_category AS ENUM ('pppoe', 'ip-static', 'hotspot', 'isolated');
	EXCEPTION
		WHEN duplicate_object THEN null;
	END $$;

	CREATE TABLE IF NOT EXISTS bandwidth_profiles (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		profile_code VARCHAR(50) UNIQUE NOT NULL,
		name VARCHAR(255) NOT NULL,
		description TEXT,
		category bandwidth_profile_category NOT NULL DEFAULT 'pppoe',
		ppp_profile_name VARCHAR(255),
		download_speed BIGINT NOT NULL DEFAULT 0,
		upload_speed BIGINT NOT NULL DEFAULT 0,
		burst_download BIGINT DEFAULT 0,
		burst_upload BIGINT DEFAULT 0,
		burst_threshold INTEGER DEFAULT 80,
		burst_time INTEGER DEFAULT 60,
		priority INTEGER DEFAULT 8 CHECK (priority >= 1 AND priority <= 8),
		queue_type VARCHAR(50) DEFAULT 'default',
		shared_users INTEGER DEFAULT 1,
		queue_name VARCHAR(255),
		price_monthly DECIMAL(12,2) DEFAULT 0,
		price_installation DECIMAL(12,2) DEFAULT 0,
		tax_rate DECIMAL(5,4) DEFAULT 0.1100,
		is_active BOOLEAN DEFAULT true,
		is_visible BOOLEAN DEFAULT true,
		sort_order INTEGER DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		deleted_at TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_bandwidth_profiles_category ON bandwidth_profiles(category);
	CREATE INDEX IF NOT EXISTS idx_bandwidth_profiles_is_active ON bandwidth_profiles(is_active);
	CREATE INDEX IF NOT EXISTS idx_bandwidth_profiles_is_visible ON bandwidth_profiles(is_visible);
	CREATE INDEX IF NOT EXISTS idx_bandwidth_profiles_code ON bandwidth_profiles(profile_code);
	CREATE INDEX IF NOT EXISTS idx_bandwidth_profiles_sort ON bandwidth_profiles(sort_order);

	-- Insert default isolated profile
	INSERT INTO bandwidth_profiles (
		profile_code,
		name,
		description,
		category,
		ppp_profile_name,
		download_speed,
		upload_speed,
		priority,
		is_active,
		is_visible,
		price_monthly,
		price_installation
	) VALUES (
		'ISOLATED',
		'Isolated Profile',
		'Limited speed profile for customers with overdue payments',
		'isolated',
		'isolated-profile',
		1024,
		512,
		8,
		true,
		false,
		0,
		0
	) ON CONFLICT (profile_code) DO NOTHING;
	`)
	if err != nil {
		return err
	}
	return nil
}

func downBandwidthProfile(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	DROP TABLE IF EXISTS bandwidth_profiles CASCADE;
	DROP TYPE IF EXISTS bandwidth_profile_category CASCADE;
	`)
	if err != nil {
		return err
	}
	return nil
}
