package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upCashTransaction, downCashTransaction)
}

func upCashTransaction(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS cash_transactions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			transaction_number VARCHAR(50) UNIQUE NOT NULL,
			transaction_date TIMESTAMP NOT NULL,
			type VARCHAR(20) NOT NULL CHECK (type IN ('income', 'expense')),
			category_id UUID NOT NULL REFERENCES cash_categories(id) ON DELETE CASCADE,
			amount DECIMAL(12,2) NOT NULL CHECK (amount >= 0),
			payment_method VARCHAR(50),
			description TEXT,
			reference_type VARCHAR(50) CHECK (reference_type IN ('payment', 'invoice', 'expense', 'other', 'none')),
			reference_id UUID,
			customer_id UUID REFERENCES customers(id) ON DELETE SET NULL,
			account_name VARCHAR(100),
			account_number VARCHAR(50),
			proof_image TEXT,
			receipt_number VARCHAR(50),
			requires_approval BOOLEAN DEFAULT false,
			approval_status VARCHAR(20) DEFAULT 'pending' CHECK (approval_status IN ('pending', 'approved', 'rejected')),
			approved_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
			approved_at TIMESTAMP,
			notes TEXT,
			processed_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
			deleted_at TIMESTAMP
		);

		CREATE INDEX idx_cash_transactions_date ON cash_transactions(transaction_date);
		CREATE INDEX idx_cash_transactions_type ON cash_transactions(type);
		CREATE INDEX idx_cash_transactions_category ON cash_transactions(category_id);
		CREATE INDEX idx_cash_transactions_customer ON cash_transactions(customer_id);
		CREATE INDEX idx_cash_transactions_reference ON cash_transactions(reference_type, reference_id);
		CREATE INDEX idx_cash_transactions_number ON cash_transactions(transaction_number);
		CREATE INDEX idx_cash_transactions_approval ON cash_transactions(approval_status);
		CREATE INDEX idx_cash_transactions_deleted_at ON cash_transactions(deleted_at);
	`)
	if err != nil {
		return err
	}

	return nil
}

func downCashTransaction(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS cash_transactions;`)
	if err != nil {
		return err
	}
	return nil
}
