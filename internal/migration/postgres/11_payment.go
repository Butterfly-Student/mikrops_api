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
	-- Create ENUM for payment method
	DO $$ BEGIN
		CREATE TYPE payment_method AS ENUM ('va', 'e-wallet', 'credit_card', 'debit_card', 'qris', 'cash', 'bank_transfer');
	EXCEPTION
		WHEN duplicate_object THEN null;
	END $$;

	-- Create ENUM for payment status
	DO $$ BEGIN
		CREATE TYPE payment_status_type AS ENUM ('pending', 'confirmed', 'rejected', 'refunded');
	EXCEPTION
		WHEN duplicate_object THEN null;
	END $$;

	CREATE TABLE IF NOT EXISTS payments (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		payment_number VARCHAR(50) UNIQUE NOT NULL,
		customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
		invoice_id UUID REFERENCES invoices(id) ON DELETE SET NULL,
		amount DECIMAL(12,2) NOT NULL DEFAULT 0,
		allocated_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
		payment_method payment_method NOT NULL,
		payment_date TIMESTAMP NOT NULL,
		bank_name VARCHAR(100),
		bank_account_number VARCHAR(100),
		bank_account_name VARCHAR(255),
		transaction_reference VARCHAR(255),
		ewallet_provider VARCHAR(50),
		ewallet_number VARCHAR(50),
		proof_image TEXT,
		receipt_number VARCHAR(100),
		status payment_status_type NOT NULL DEFAULT 'pending',
		processed_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
		processed_at TIMESTAMP,
		rejection_reason TEXT,
		refund_amount DECIMAL(12,2) DEFAULT 0,
		refund_date TIMESTAMP,
		refund_reason TEXT,
		refunded_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
		notes TEXT,
		created_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		deleted_at TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_payments_customer ON payments(customer_id);
	CREATE INDEX IF NOT EXISTS idx_payments_invoice ON payments(invoice_id);
	CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status);
	CREATE INDEX IF NOT EXISTS idx_payments_payment_date ON payments(payment_date);
	CREATE INDEX IF NOT EXISTS idx_payments_method ON payments(payment_method);
	CREATE INDEX IF NOT EXISTS idx_payments_number ON payments(payment_number);
	CREATE INDEX IF NOT EXISTS idx_payments_transaction_ref ON payments(transaction_reference);
	CREATE INDEX IF NOT EXISTS idx_payments_created_by ON payments(created_by);
	CREATE INDEX IF NOT EXISTS idx_payments_processed_by ON payments(processed_by);
	CREATE INDEX IF NOT EXISTS idx_payments_deleted_at ON payments(deleted_at);
	`)
	if err != nil {
		return err
	}
	return nil
}

func downPayment(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	DROP TABLE IF EXISTS payments CASCADE;
	DROP TYPE IF EXISTS payment_method CASCADE;
	DROP TYPE IF EXISTS payment_status_type CASCADE;
	`)
	if err != nil {
		return err
	}
	return nil
}
