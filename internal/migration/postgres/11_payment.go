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
	_, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS payments (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			payment_number VARCHAR(50) UNIQUE NOT NULL,
			customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
			invoice_id UUID REFERENCES invoices(id) ON DELETE SET NULL,
			amount DECIMAL(12,2) NOT NULL CHECK (amount >= 0),
			allocated_amount DECIMAL(12,2) DEFAULT 0 CHECK (allocated_amount >= 0),
			payment_method VARCHAR(20) NOT NULL CHECK (payment_method IN ('va', 'e-wallet', 'credit_card', 'debit_card', 'check')),
			payment_date TIMESTAMP NOT NULL,
			bank_name VARCHAR(100),
			bank_account_number VARCHAR(50),
			bank_account_name VARCHAR(100),
			transaction_reference VARCHAR(100),
			ewallet_provider VARCHAR(20) CHECK (ewallet_provider IS NULL OR ewallet_provider IN ('gopay', 'ovo', 'dana')),
			ewallet_number VARCHAR(30),
			proof_image TEXT,
			receipt_number VARCHAR(50),
			status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'confirmed', 'rejected', 'refunded')),
			processed_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
			processed_at TIMESTAMP,
			rejection_reason TEXT,
			refund_amount DECIMAL(12,2) DEFAULT 0 CHECK (refund_amount >= 0),
			refund_date TIMESTAMP,
			refund_reason TEXT,
			refunded_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
			notes TEXT,
			created_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
			deleted_at TIMESTAMP
		);

		CREATE INDEX idx_payments_customer ON payments(customer_id);
		CREATE INDEX idx_payments_invoice ON payments(invoice_id);
		CREATE INDEX idx_payments_status ON payments(status);
		CREATE INDEX idx_payments_date ON payments(payment_date);
		CREATE INDEX idx_payments_method ON payments(payment_method);
		CREATE INDEX idx_payments_number ON payments(payment_number);
		CREATE INDEX idx_payments_deleted_at ON payments(deleted_at);
	`)
	if err != nil {
		return err
	}
	return nil
}

func downPayment(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS payments;`)
	if err != nil {
		return err
	}
	return nil
}
