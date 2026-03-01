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
	CREATE TABLE IF NOT EXISTS customers (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		customer_code VARCHAR(20) UNIQUE NOT NULL,
		full_name VARCHAR(100) NOT NULL,
		phone VARCHAR(20),
		email VARCHAR(100),
		address TEXT,
		coordinates GEOMETRY(POINT, 4326),
		id_card_number VARCHAR(30),
		status VARCHAR(20) CHECK (status IN ('active', 'suspended', 'terminated', 'pending')) DEFAULT 'pending',
		notes TEXT,
		created_by UUID REFERENCES admin_users(id),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_customers_code ON customers(customer_code);
	CREATE INDEX IF NOT EXISTS idx_customers_status ON customers(status);
	CREATE INDEX IF NOT EXISTS idx_customers_phone ON customers(phone);
	CREATE INDEX IF NOT EXISTS idx_customers_email ON customers(email) WHERE email IS NOT NULL;

	COMMENT ON TABLE customers IS 'Data utama pelanggan ISP';
	COMMENT ON COLUMN customers.coordinates IS 'Koordinat GPS untuk peta jaringan (PostGIS)';
	`)
	return err
}

func downCustomers(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS customers CASCADE;`)
	return err
}
