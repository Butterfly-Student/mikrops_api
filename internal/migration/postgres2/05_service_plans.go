package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upServicePlans, downServicePlans)
}

func upServicePlans(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	CREATE TABLE IF NOT EXISTS service_plans (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		plan_code VARCHAR(20) UNIQUE NOT NULL,
		plan_name VARCHAR(100) NOT NULL,
		service_type VARCHAR(20) NOT NULL CHECK (service_type IN ('pppoe', 'hotspot', 'static_ip', 'vpn')),
		price NUMERIC(12,2) NOT NULL,
		billing_cycle VARCHAR(20) CHECK (billing_cycle IN ('daily', 'weekly', 'monthly', 'yearly')) DEFAULT 'monthly',
		speed_up INTEGER NOT NULL,
		speed_down INTEGER NOT NULL,
		burst_up INTEGER DEFAULT 0,
		burst_down INTEGER DEFAULT 0,
		burst_threshold INTEGER DEFAULT 0,
		burst_time INTEGER DEFAULT 0,
		quota_gb INTEGER,
		mikrotik_profile VARCHAR(100),
		address_pool VARCHAR(100),
		description TEXT,
		is_active BOOLEAN DEFAULT true,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_service_plans_code ON service_plans(plan_code);
	CREATE INDEX IF NOT EXISTS idx_service_plans_type ON service_plans(service_type);
	CREATE INDEX IF NOT EXISTS idx_service_plans_active ON service_plans(is_active);

	COMMENT ON TABLE service_plans IS 'Master paket layanan semua jenis';
	COMMENT ON COLUMN service_plans.quota_gb IS 'NULL = unlimited';
	`)
	return err
}

func downServicePlans(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS service_plans CASCADE;`)
	return err
}
