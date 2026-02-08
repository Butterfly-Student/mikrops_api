package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upInvoice, downInvoice)
}

func upInvoice(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`CREATE TABLE IF NOT EXISTS invoices (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		tenant_id UUID NOT NULL,
		customer_id UUID NOT NULL,
		subscription_id UUID NOT NULL,
		invoice_number VARCHAR(100) UNIQUE NOT NULL,
		amount BIGINT NOT NULL,
		tax_amount BIGINT DEFAULT 0,
		total_amount BIGINT NOT NULL,
		status VARCHAR(50) DEFAULT 'unpaid',
		due_date TIMESTAMP NOT NULL,
		period_start TIMESTAMP NOT NULL,
		period_end TIMESTAMP NOT NULL,
		notes TEXT,
		paid_at TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
		FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE CASCADE,
		FOREIGN KEY (subscription_id) REFERENCES subscriptions(id) ON DELETE CASCADE
	);
	
	CREATE INDEX IF NOT EXISTS idx_invoices_tenant_id ON invoices(tenant_id);
	CREATE INDEX IF NOT EXISTS idx_invoices_customer_id ON invoices(customer_id);
	CREATE INDEX IF NOT EXISTS idx_invoices_subscription_id ON invoices(subscription_id);
	CREATE INDEX IF NOT EXISTS idx_invoices_status ON invoices(status);
	CREATE INDEX IF NOT EXISTS idx_invoices_invoice_number ON invoices(invoice_number);
	CREATE INDEX IF NOT EXISTS idx_invoices_due_date ON invoices(due_date);`)
	if err != nil {
		return err
	}
	return nil
}

func downInvoice(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS invoices CASCADE;`)
	if err != nil {
		return err
	}
	return nil
}
