package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upCustomersBase, downCustomersBase)
}

// upCustomersBase creates the customers table (identitas saja — PPPoE/layanan ada di subscriptions).
func upCustomersBase(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	-- ============================================
	-- CUSTOMERS TABLE (identitas pelanggan)
	-- ============================================
	CREATE TABLE IF NOT EXISTS customers (
		id            UUID      PRIMARY KEY DEFAULT gen_random_uuid(),
		customer_code VARCHAR(50) UNIQUE NOT NULL,

		-- Identitas
		full_name  VARCHAR(100) NOT NULL,
		email      VARCHAR(100) UNIQUE,
		phone      VARCHAR(20)  NOT NULL,
		id_card_number VARCHAR(30), -- NIK KTP

		-- Alamat
		address   TEXT,
		latitude  DECIMAL(10,8),
		longitude DECIMAL(11,8),

		-- Status Pelanggan
		status            VARCHAR(20) NOT NULL
		                  CHECK (status IN ('pending','active','suspended','isolated','terminated'))
		                  DEFAULT 'pending',
		activation_date   DATE,
		installation_date DATE,
		termination_date  DATE,

		-- Portal Self-Service
		portal_password   VARCHAR(255), -- bcrypt hash untuk login portal
		portal_last_login TIMESTAMPTZ,


		-- Billing otomatis
		auto_isolate      BOOLEAN DEFAULT true,
		grace_period_days INTEGER DEFAULT 3,

		notes TEXT,
		tags  JSONB,

		created_by UUID,
		updated_by UUID,
		created_at TIMESTAMP   DEFAULT CURRENT_TIMESTAMP NOT NULL,
		updated_at TIMESTAMP   DEFAULT CURRENT_TIMESTAMP NOT NULL,
		deleted_at TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_customers_code    ON customers(customer_code);
	CREATE INDEX IF NOT EXISTS idx_customers_status  ON customers(status)  WHERE deleted_at IS NULL;
	CREATE INDEX IF NOT EXISTS idx_customers_phone   ON customers(phone);
	CREATE INDEX IF NOT EXISTS idx_customers_email   ON customers(email)   WHERE email IS NOT NULL;
	CREATE INDEX IF NOT EXISTS idx_customers_deleted ON customers(deleted_at);

	COMMENT ON TABLE  customers                IS 'Data identitas pelanggan ISP (konfigurasi layanan di tabel subscriptions)';
	COMMENT ON COLUMN customers.status          IS 'pending=menunggu instalasi, active=aktif, suspended=ditangguhkan manual, isolated=diisolasi karena telat bayar, terminated=berakhir';
	COMMENT ON COLUMN customers.auto_isolate    IS 'Aktifkan auto-isolasi jika tagihan melewati grace period';
	COMMENT ON COLUMN customers.grace_period_days IS 'Hari toleransi setelah due date sebelum di-isolasi';
	COMMENT ON COLUMN customers.previous_profile_id IS 'Profil bandwidth sebelum isolasi, dikembalikan saat isolasi dicabut';
	`)
	return err
}

func downCustomersBase(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS customers CASCADE;`)
	return err
}
