package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigration(UpAlterInvoicesAddXenditFields, DownAlterInvoicesAddXenditFields)
}

func UpAlterInvoicesAddXenditFields(tx *sql.Tx) error {
	// Add external_id column
	_, err := tx.Exec(`ALTER TABLE invoices ADD COLUMN external_id VARCHAR(255);`)
	if err != nil {
		return err
	}

	// Add invoice_url column
	_, err = tx.Exec(`ALTER TABLE invoices ADD COLUMN invoice_url VARCHAR(255);`)
	if err != nil {
		return err
	}

	// Add index for external_id for faster lookups
	_, err = tx.Exec(`CREATE INDEX idx_invoices_external_id ON invoices(external_id);`)
	if err != nil {
		return err
	}

	return nil
}

func DownAlterInvoicesAddXenditFields(tx *sql.Tx) error {
	// Remove index
	_, err := tx.Exec(`DROP INDEX IF EXISTS idx_invoices_external_id;`)
	if err != nil {
		return err
	}

	// Remove invoice_url column
	_, err = tx.Exec(`ALTER TABLE invoices DROP COLUMN invoice_url;`)
	if err != nil {
		return err
	}

	// Remove external_id column
	_, err = tx.Exec(`ALTER TABLE invoices DROP COLUMN external_id;`)
	if err != nil {
		return err
	}

	return nil
}
