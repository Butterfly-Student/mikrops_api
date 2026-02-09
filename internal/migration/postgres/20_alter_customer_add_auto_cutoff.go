package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upAlterCustomerAddAutoCutoff, downAlterCustomerAddAutoCutoff)
}

func upAlterCustomerAddAutoCutoff(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`ALTER TABLE customers ADD COLUMN IF NOT EXISTS auto_cutoff BOOLEAN DEFAULT true;`)
	if err != nil {
		return err
	}
	return nil
}

func downAlterCustomerAddAutoCutoff(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`ALTER TABLE customers DROP COLUMN IF EXISTS auto_cutoff;`)
	if err != nil {
		return err
	}
	return nil
}
