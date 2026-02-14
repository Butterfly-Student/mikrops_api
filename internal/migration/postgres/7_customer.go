package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upCustomer, downCustomer)
}

func upCustomer(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	-- Create ENUM for customer status
	DO $$ BEGIN
		CREATE TYPE customer_status AS ENUM ('pending', 'active', 'suspended', 'isolated', 'terminated');
	EXCEPTION
		WHEN duplicate_object THEN null;
	END $$;

	-- Create ENUM for PPP service type
	DO $$ BEGIN
		CREATE TYPE ppp_service_type AS ENUM ('pppoe', 'pptp', 'l2tp', 'ovpn', 'any');
	EXCEPTION
		WHEN duplicate_object THEN null;
	END $$;

	-- Create ENUM for billing cycle
	DO $$ BEGIN
		CREATE TYPE billing_cycle_type AS ENUM ('monthly', 'quarterly', 'semi-annually', 'yearly');
	EXCEPTION
		WHEN duplicate_object THEN null;
	END $$;

	-- Enable PostGIS extension for geography support
	CREATE EXTENSION IF NOT EXISTS postgis;

	CREATE TABLE IF NOT EXISTS customers (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		customer_code VARCHAR(50) UNIQUE NOT NULL,
		full_name VARCHAR(255) NOT NULL,
		email VARCHAR(255),
		phone VARCHAR(50),
		address TEXT,
		coordinates GEOGRAPHY(POINT, 4326),
		status customer_status NOT NULL DEFAULT 'pending',
		activation_date DATE,
		installation_date DATE,
		termination_date DATE,
		expiry_date DATE,
		router_id UUID REFERENCES mikrotik_routers(id) ON DELETE SET NULL,
		ppp_secret_name VARCHAR(255) UNIQUE,
		ppp_secret_password TEXT,
		ppp_service ppp_service_type DEFAULT 'pppoe',
		static_ip VARCHAR(50),
		mac_address VARCHAR(50),
		profile_id UUID REFERENCES bandwidth_profiles(id) ON DELETE SET NULL,
		billing_cycle billing_cycle_type DEFAULT 'monthly',
		billing_day INTEGER DEFAULT 1 CHECK (billing_day >= 1 AND billing_day <= 31),
		payment_method_preference VARCHAR(50),
		auto_isolate BOOLEAN DEFAULT true,
		grace_period_days INTEGER DEFAULT 3,
		notes TEXT,
		tags JSONB DEFAULT '[]'::jsonb,
		created_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
		updated_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		deleted_at TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_customers_status ON customers(status);
	CREATE INDEX IF NOT EXISTS idx_customers_router ON customers(router_id);
	CREATE INDEX IF NOT EXISTS idx_customers_profile ON customers(profile_id);
	CREATE INDEX IF NOT EXISTS idx_customers_code ON customers(customer_code);
	CREATE INDEX IF NOT EXISTS idx_customers_phone ON customers(phone);
	CREATE INDEX IF NOT EXISTS idx_customers_email ON customers(email);
	CREATE INDEX IF NOT EXISTS idx_customers_expiry ON customers(expiry_date);
	CREATE INDEX IF NOT EXISTS idx_customers_ppp_secret ON customers(ppp_secret_name);
	CREATE INDEX IF NOT EXISTS idx_customers_coordinates ON customers USING GIST(coordinates);
	CREATE INDEX IF NOT EXISTS idx_customers_created_by ON customers(created_by);
	CREATE INDEX IF NOT EXISTS idx_customers_updated_by ON customers(updated_by);
	CREATE INDEX IF NOT EXISTS idx_customers_deleted_at ON customers(deleted_at);
	`)
	if err != nil {
		return err
	}
	return nil
}

func downCustomer(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	DROP TABLE IF EXISTS customers CASCADE;
	DROP TYPE IF EXISTS customer_status CASCADE;
	DROP TYPE IF EXISTS ppp_service_type CASCADE;
	DROP TYPE IF EXISTS billing_cycle_type CASCADE;
	`)
	if err != nil {
		return err
	}
	return nil
}
