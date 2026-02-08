package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upPayment, downPayment)
}

func upPayment(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`CREATE TABLE IF NOT EXISTS payments (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		tenant_id UUID NOT NULL,
		invoice_id UUID NOT NULL,
		payment_method_id UUID NOT NULL,
		amount BIGINT NOT NULL,
		payment_date TIMESTAMP NOT NULL,
		proof_url VARCHAR(500),
		status VARCHAR(50) DEFAULT 'pending',
		verified_by UUID,
		verified_at TIMESTAMP,
		notes TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
		FOREIGN KEY (invoice_id) REFERENCES invoices(id) ON DELETE CASCADE,
		FOREIGN KEY (payment_method_id) REFERENCES payment_methods(id) ON DELETE RESTRICT,
		FOREIGN KEY (verified_by) REFERENCES staffs(id) ON DELETE SET NULL
	);
	
	CREATE INDEX IF NOT EXISTS idx_payments_tenant_id ON payments(tenant_id);
	CREATE INDEX IF NOT EXISTS idx_payments_invoice_id ON payments(invoice_id);
	CREATE INDEX IF NOT EXISTS idx_payments_payment_method_id ON payments(payment_method_id);
	CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status);
	CREATE INDEX IF NOT EXISTS idx_payments_verified_by ON payments(verified_by);`)
	if err != nil {
		return err
	}
	return nil
}

func downPayment(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS payments CASCADE;`)
	if err != nil {
		return err
	}
	return nil
}
