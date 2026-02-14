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
					id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
					refund_number VARCHAR(255) UNIQUE NOT NULL,
					payment_id UUID NOT NULL,
					invoice_id UUID,
					customer_id UUID NOT NULL,
					refund_amount DECIMAL(12,2) NOT NULL,
					refund_type VARCHAR(50) NOT NULL CHECK (refund_type IN ('full', 'partial')),
					refund_reason TEXT,
					refund_method VARCHAR(50) CHECK (refund_method IN ('original', 'bank_transfer', 'cash')),
					bank_name VARCHAR(255),
					bank_account_name VARCHAR(255),
					bank_account_number VARCHAR(255),
					status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'processed', 'completed', 'failed')),
					approved_by UUID,
					approved_at TIMESTAMP,
					processed_by UUID,
					processed_at TIMESTAMP,
					rejection_reason TEXT,
					notes TEXT,
					xendit_refund_id VARCHAR(255),
					created_by UUID,
					created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
					deleted_at TIMESTAMP,
					FOREIGN KEY (payment_id) REFERENCES payments(id) ON DELETE CASCADE,
					FOREIGN KEY (invoice_id) REFERENCES invoices(id) ON DELETE SET NULL,
					FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE CASCADE,
					FOREIGN KEY (approved_by) REFERENCES users(id) ON DELETE SET NULL,
					FOREIGN KEY (processed_by) REFERENCES users(id) ON DELETE SET NULL,
					FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL
				);

				CREATE INDEX IF NOT EXISTS idx_refunds_payment ON refunds(payment_id);
				CREATE INDEX IF NOT EXISTS idx_refunds_invoice ON refunds(invoice_id);
				CREATE INDEX IF NOT EXISTS idx_refunds_customer ON refunds(customer_id);
				CREATE INDEX IF NOT EXISTS idx_refunds_status ON refunds(status);
				CREATE INDEX IF NOT EXISTS idx_refunds_type ON refunds(refund_type);
				CREATE INDEX IF NOT EXISTS idx_refunds_number ON refunds(refund_number);
				CREATE INDEX IF NOT EXISTS idx_refunds_created_at ON refunds(created_at);
				CREATE INDEX IF NOT EXISTS idx_refunds_deleted_at ON refunds(deleted_at);

				COMMENT ON TABLE refunds IS 'Stores payment refund information';
				COMMENT ON COLUMN refunds.refund_type IS 'Type of refund: full or partial';
				COMMENT ON COLUMN refunds.refund_method IS 'Method of refund: original payment method, bank transfer, or cash';
				COMMENT ON COLUMN refunds.status IS 'Refund status workflow';
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
