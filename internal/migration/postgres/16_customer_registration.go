package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upCustomerRegistration, downCustomerRegistration)
}

func upCustomerRegistration(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`CREATE TABLE IF NOT EXISTS customer_registrations (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		tenant_id UUID NOT NULL,
		customer_id UUID,
		full_name VARCHAR(255) NOT NULL,
		email VARCHAR(255),
		phone VARCHAR(50),
		address TEXT,
		requested_package_id UUID NOT NULL,
		requested_nas_id UUID,
		status VARCHAR(50) DEFAULT 'pending',
		rejection_reason TEXT,
		approved_by UUID,
		approved_at TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
		FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE SET NULL,
		FOREIGN KEY (requested_package_id) REFERENCES internet_packages(id) ON DELETE RESTRICT,
		FOREIGN KEY (requested_nas_id) REFERENCES nas(id) ON DELETE SET NULL,
		FOREIGN KEY (approved_by) REFERENCES staffs(id) ON DELETE SET NULL
	);

	CREATE INDEX IF NOT EXISTS idx_customer_registrations_tenant_id ON customer_registrations(tenant_id);
	CREATE INDEX IF NOT EXISTS idx_customer_registrations_status ON customer_registrations(status);
	CREATE INDEX IF NOT EXISTS idx_customer_registrations_customer_id ON customer_registrations(customer_id);`)
	if err != nil {
		return err
	}
	return nil
}

func downCustomerRegistration(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS customer_registrations CASCADE;`)
	if err != nil {
		return err
	}
	return nil
}
