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
	_, err := tx.Exec(`CREATE TABLE customers (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		customer_code VARCHAR(50) UNIQUE NOT NULL,
		full_name VARCHAR(100) NOT NULL,
		email VARCHAR(100) UNIQUE,
		phone VARCHAR(20) NOT NULL,
		address TEXT,
		latitude DECIMAL(10, 8),
		longitude DECIMAL(11, 8),
		status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'active', 'suspended', 'isolated', 'terminated')),
		activation_date DATE,
		installation_date DATE,
		termination_date DATE,
		expiry_date DATE,
		router_id UUID REFERENCES mikrotik_routers(id) ON DELETE RESTRICT,
		ppp_secret_name VARCHAR(100) UNIQUE,
		ppp_secret_password TEXT,
		ppp_service VARCHAR(20) DEFAULT 'pppoe' CHECK (ppp_service IN ('pppoe', 'pptp', 'l2tp', 'ovpn')),
		static_ip VARCHAR(45),
		mac_address VARCHAR(17),
		profile_id UUID REFERENCES bandwidth_profiles(id) ON DELETE RESTRICT,
		billing_cycle VARCHAR(20) DEFAULT 'monthly' CHECK (billing_cycle IN ('monthly', 'quarterly', 'yearly')),
		billing_day INTEGER DEFAULT 1 CHECK (billing_day BETWEEN 1 AND 31),
		payment_method_preference VARCHAR(20) CHECK (payment_method_preference IN ('cash', 'transfer', 'e-wallet', 'auto-debit')),
		auto_isolate BOOLEAN DEFAULT true,
		grace_period_days INTEGER DEFAULT 3,
		notes TEXT,
		tags TEXT[] DEFAULT '{}',
		created_by INTEGER REFERENCES users(id),
		updated_by INTEGER REFERENCES users(id),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		deleted_at TIMESTAMP
	);`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE INDEX idx_customers_status ON customers(status) WHERE deleted_at IS NULL;`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE INDEX idx_customers_router ON customers(router_id) WHERE deleted_at IS NULL;`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE INDEX idx_customers_profile ON customers(profile_id) WHERE deleted_at IS NULL;`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE INDEX idx_customers_code ON customers(customer_code);`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE INDEX idx_customers_phone ON customers(phone);`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE INDEX idx_customers_email ON customers(email) WHERE email IS NOT NULL;`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE INDEX idx_customers_expiry ON customers(expiry_date) WHERE deleted_at IS NULL;`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE INDEX idx_customers_ppp_secret ON customers(ppp_secret_name) WHERE deleted_at IS NULL;`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE INDEX idx_customers_latitude ON customers(latitude) WHERE latitude IS NOT NULL;`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE INDEX idx_customers_longitude ON customers(longitude) WHERE longitude IS NOT NULL;`)
	if err != nil {
		return err
	}

	return nil
}

func downCustomer(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS customers CASCADE;`)
	return err
}
