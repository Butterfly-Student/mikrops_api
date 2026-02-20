package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upCustomers, downCustomers)
}

func upCustomers(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	-- ============================================
	-- CUSTOMERS TABLE
	-- ============================================
	CREATE TABLE IF NOT EXISTS customers (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		customer_code VARCHAR(50) UNIQUE NOT NULL,

		-- Personal Information
		full_name VARCHAR(100) NOT NULL,
		email VARCHAR(100) UNIQUE,
		phone VARCHAR(20) NOT NULL,

		-- Address Information
		address TEXT,
		latitude DECIMAL(10,8),
		longitude DECIMAL(11,8),

		-- Service Information
		status VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'active', 'suspended', 'isolated', 'terminated')) DEFAULT 'pending',
		activation_date DATE,
		installation_date DATE,
		termination_date DATE,
		expiry_date DATE, -- tanggal jatuh tempo berlangganan

		-- Mikrotik Configuration
		router_id UUID REFERENCES mikrotik_routers(id) ON DELETE RESTRICT,
		ppp_secret_name VARCHAR(100) UNIQUE, -- username PPP di Mikrotik
		ppp_secret_password VARCHAR(255), -- encrypted
		ppp_service VARCHAR(20) DEFAULT 'pppoe' CHECK (ppp_service IN ('pppoe', 'pptp', 'l2tp', 'ovpn')),
		static_ip VARCHAR(45), -- optional static IP
		mac_address VARCHAR(17), -- optional MAC binding

		-- Package Information
		profile_id UUID REFERENCES bandwidth_profiles(id) ON DELETE RESTRICT,

		-- Billing Information
		billing_cycle VARCHAR(20) DEFAULT 'monthly' CHECK (billing_cycle IN ('monthly', 'quarterly', 'yearly')),
		billing_day INTEGER DEFAULT 1 CHECK (billing_day BETWEEN 1 AND 31), -- tanggal tagihan
		payment_method_preference VARCHAR(20) CHECK (payment_method_preference IN ('cash', 'transfer', 'e-wallet', 'auto-debit')),
		auto_isolate BOOLEAN DEFAULT true,
		grace_period_days INTEGER DEFAULT 3, -- hari toleransi setelah jatuh tempo

		notes TEXT,
		tags JSONB, -- untuk labeling/categorization

		created_by UUID, -- FK to users
		updated_by UUID, -- FK to users
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		deleted_at TIMESTAMP -- soft delete
	);

	-- Indexes
	CREATE INDEX IF NOT EXISTS idx_customers_status ON customers(status) WHERE deleted_at IS NULL;
	CREATE INDEX IF NOT EXISTS idx_customers_router ON customers(router_id) WHERE deleted_at IS NULL;
	CREATE INDEX IF NOT EXISTS idx_customers_profile ON customers(profile_id) WHERE deleted_at IS NULL;
	CREATE INDEX IF NOT EXISTS idx_customers_code ON customers(customer_code);
	CREATE INDEX IF NOT EXISTS idx_customers_phone ON customers(phone);
	CREATE INDEX IF NOT EXISTS idx_customers_email ON customers(email) WHERE email IS NOT NULL;
	CREATE INDEX IF NOT EXISTS idx_customers_expiry ON customers(expiry_date) WHERE deleted_at IS NULL;
	CREATE INDEX IF NOT EXISTS idx_customers_ppp_secret ON customers(ppp_secret_name) WHERE deleted_at IS NULL;
	CREATE INDEX IF NOT EXISTS idx_customers_deleted ON customers(deleted_at);

	-- Comments
	COMMENT ON TABLE customers IS 'Data pelanggan internet';
	COMMENT ON COLUMN customers.status IS 'pending=menunggu instalasi, active=aktif berlangganan, suspended=ditangguhkan manual, isolated=diisolasi karena telat bayar, terminated=berakhir';
	COMMENT ON COLUMN customers.grace_period_days IS 'Jumlah hari grace period sebelum auto isolate';
	COMMENT ON COLUMN customers.auto_isolate IS 'Flag untuk enable/disable auto isolate untuk customer ini';
	`)
	if err != nil {
		return err
	}
	return nil
}

func downCustomers(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS customers CASCADE;`)
	if err != nil {
		return err
	}
	return nil
}
