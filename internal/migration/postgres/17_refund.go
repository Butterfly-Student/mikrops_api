package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upRefund, downRefund)
}

func upRefund(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
				CREATE TABLE IF NOT EXISTS refunds (
					id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					refund_number VARCHAR(255) UNIQUE NOT NULL,
					payment_id UUID NOT NULL REFERENCES payments(id) ON DELETE CASCADE,
					invoice_id UUID REFERENCES invoices(id) ON DELETE SET NULL,
					customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
					refund_amount DECIMAL(12,2) NOT NULL CHECK (refund_amount >= 0),
					refund_type VARCHAR(50) NOT NULL CHECK (refund_type IN ('full', 'partial')),
					refund_reason TEXT,
					refund_method VARCHAR(50) CHECK (refund_method IN ('original', 'bank_transfer', 'cash')),
					bank_name VARCHAR(255),
					bank_account_name VARCHAR(255),
					bank_account_number VARCHAR(255),
					status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'processed', 'completed', 'failed')),
					approved_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
					approved_at TIMESTAMP,
					processed_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
					processed_at TIMESTAMP,
					rejection_reason TEXT,
					notes TEXT,
					xendit_refund_id VARCHAR(255),
					created_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
					updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
					deleted_at TIMESTAMP
				);

				CREATE INDEX IF NOT EXISTS idx_refunds_payment ON refunds(payment_id);
				CREATE INDEX IF NOT EXISTS idx_refunds_invoice ON refunds(invoice_id);
				CREATE INDEX IF NOT EXISTS idx_refunds_customer ON refunds(customer_id);
				CREATE INDEX IF NOT EXISTS idx_refunds_status ON refunds(status) WHERE deleted_at IS NULL;
				CREATE INDEX IF NOT EXISTS idx_refunds_type ON refunds(refund_type);
				CREATE INDEX IF NOT EXISTS idx_refunds_number ON refunds(refund_number);
				CREATE INDEX IF NOT EXISTS idx_refunds_created_at ON refunds(created_at);
				CREATE INDEX IF NOT EXISTS idx_refunds_deleted_at ON refunds(deleted_at);
	`)
	if err != nil {
		return err
	}
	return nil
}

func downRefund(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS refunds CASCADE;`)
	if err != nil {
		return err
	}
	return nil
}
