package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upCashCategory, downCashCategory)
}

func upCashCategory(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS cash_categories (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			code VARCHAR(50) UNIQUE NOT NULL,
			name VARCHAR(200) NOT NULL,
			type VARCHAR(20) NOT NULL CHECK (type IN ('income', 'expense')),
			parent_category_id UUID REFERENCES cash_categories(id) ON DELETE CASCADE,
			description TEXT,
			is_system BOOLEAN NOT NULL DEFAULT false,
			is_active BOOLEAN NOT NULL DEFAULT true,
			sort_order INTEGER DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
		);

		CREATE INDEX idx_cash_categories_type ON cash_categories(type);
		CREATE INDEX idx_cash_categories_parent ON cash_categories(parent_category_id);
	`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO cash_categories (code, name, type, description, is_system, sort_order) VALUES
		('INC-SUB', 'Subscription Revenue', 'income', 'Monthly subscription fees from customers', true, 1),
		('INC-INS', 'Installation Revenue', 'income', 'One-time installation fees', true, 2),
		('INC-OTH', 'Other Income', 'income', 'Miscellaneous income', true, 3),
		('EXP-OPR', 'Operational Expense', 'expense', 'Daily operational costs', true, 1),
		('EXP-EQP', 'Equipment Purchase', 'expense', 'Network equipment and hardware', true, 2),
		('EXP-SAL', 'Salary', 'expense', 'Employee salaries and wages', true, 3),
		('EXP-UTL', 'Utilities', 'expense', 'Electricity, internet, and other utilities', true, 4),
		('EXP-MNT', 'Maintenance', 'expense', 'Maintenance and repair costs', true, 5),
		('EXP-OTH', 'Other Expense', 'expense', 'Miscellaneous expenses', true, 6);
	`)
	if err != nil {
		return err
	}

	return nil
}

func downCashCategory(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS cash_categories;`)
	if err != nil {
		return err
	}
	return nil
}
