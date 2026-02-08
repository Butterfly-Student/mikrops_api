package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upPaymentMethod, downPaymentMethod)
}

func upPaymentMethod(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`CREATE TABLE IF NOT EXISTS payment_methods (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		tenant_id UUID NOT NULL,
		name VARCHAR(255) NOT NULL,
		type VARCHAR(50) NOT NULL,
		account_name VARCHAR(255),
		account_number VARCHAR(100),
		bank_name VARCHAR(255),
		instructions TEXT,
		is_active BOOLEAN DEFAULT true,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
	);
	
	CREATE INDEX IF NOT EXISTS idx_payment_methods_tenant_id ON payment_methods(tenant_id);
	CREATE INDEX IF NOT EXISTS idx_payment_methods_type ON payment_methods(type);`)
	if err != nil {
		return err
	}
	return nil
}

func downPaymentMethod(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS payment_methods CASCADE;`)
	if err != nil {
		return err
	}
	return nil
}
