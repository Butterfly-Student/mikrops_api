package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upAlterCustomerRemovePppoeFields, downAlterCustomerRemovePppoeFields)
}

func upAlterCustomerRemovePppoeFields(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`ALTER TABLE customers DROP COLUMN IF EXISTS pppoe_username;
	ALTER TABLE customers DROP COLUMN IF EXISTS pppoe_password;`)
	if err != nil {
		return err
	}
	return nil
}

func downAlterCustomerRemovePppoeFields(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`ALTER TABLE customers ADD COLUMN IF NOT EXISTS pppoe_username VARCHAR(100);
	ALTER TABLE customers ADD COLUMN IF NOT EXISTS pppoe_password VARCHAR(255);
	CREATE INDEX IF NOT EXISTS idx_customers_pppoe_username ON customers(pppoe_username);`)
	if err != nil {
		return err
	}
	return nil
}
