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
	_, err := tx.Exec(`CREATE TABLE IF NOT EXISTS customers (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		tenant_id UUID NOT NULL,
		full_name VARCHAR(255) NOT NULL,
		email VARCHAR(255),
		phone VARCHAR(50),
		address TEXT,
		identity_number VARCHAR(50),
		username VARCHAR(100),
		password_hash VARCHAR(255),
		pppoe_username VARCHAR(100),
		pppoe_password VARCHAR(255),
		static_ip VARCHAR(45),
		nas_id UUID,
		is_active BOOLEAN DEFAULT true,
		registered_at TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
		FOREIGN KEY (nas_id) REFERENCES nas(id) ON DELETE SET NULL
	);
	
	CREATE INDEX IF NOT EXISTS idx_customers_tenant_id ON customers(tenant_id);
	CREATE INDEX IF NOT EXISTS idx_customers_email ON customers(email);
	CREATE INDEX IF NOT EXISTS idx_customers_username ON customers(username);
	CREATE INDEX IF NOT EXISTS idx_customers_pppoe_username ON customers(pppoe_username);
	CREATE INDEX IF NOT EXISTS idx_customers_nas_id ON customers(nas_id);`)
	if err != nil {
		return err
	}
	return nil
}

func downCustomer(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS customers CASCADE;`)
	if err != nil {
		return err
	}
	return nil
}
