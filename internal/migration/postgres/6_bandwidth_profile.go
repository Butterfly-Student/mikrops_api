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
	_, err := tx.Exec(`CREATE TABLE bandwidth_profiles (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		profile_code VARCHAR(50) UNIQUE NOT NULL,
		name VARCHAR(100) NOT NULL,
		description TEXT,
		category VARCHAR(20) NOT NULL CHECK (category IN ('pppoe', 'ip-static', 'hotspot', 'isolated')),
		ppp_profile_name VARCHAR(100) NOT NULL,
		download_speed BIGINT NOT NULL,
		upload_speed BIGINT NOT NULL,
		burst_download BIGINT,
		burst_upload BIGINT,
		burst_threshold INTEGER DEFAULT 80,
		burst_time INTEGER DEFAULT 8,
		priority INTEGER DEFAULT 8 CHECK (priority BETWEEN 1 AND 8),
		queue_type VARCHAR(20) DEFAULT 'default',
		shared_users INTEGER DEFAULT 1,
		queue_name VARCHAR(20),
		price_monthly DECIMAL(12,2) NOT NULL,
		price_installation DECIMAL(12,2) DEFAULT 0,
		tax_rate DECIMAL(5,4) DEFAULT 0.11,
		is_active BOOLEAN DEFAULT true,
		is_visible BOOLEAN DEFAULT true,
		sort_order INTEGER DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		deleted_at TIMESTAMP
	);`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE INDEX idx_bandwidth_profiles_category ON bandwidth_profiles(category);`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE INDEX idx_bandwidth_profiles_is_active ON bandwidth_profiles(is_active);`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE INDEX idx_bandwidth_profiles_is_visible ON bandwidth_profiles(is_visible);`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE INDEX idx_bandwidth_profiles_code ON bandwidth_profiles(profile_code);`)
	if err != nil {
		return err
	}

	return nil
}

func downBandwidthProfile(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS bandwidth_profiles CASCADE;`)
	return err
}
