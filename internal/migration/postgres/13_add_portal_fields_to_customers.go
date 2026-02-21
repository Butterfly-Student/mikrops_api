package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upAddPortalFieldsToCustomers, downAddPortalFieldsToCustomers)
}

func upAddPortalFieldsToCustomers(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
		ALTER TABLE customers
		ADD COLUMN IF NOT EXISTS portal_password VARCHAR(255),
		ADD COLUMN IF NOT EXISTS portal_last_login TIMESTAMPTZ;
	`)
	return err
}

func downAddPortalFieldsToCustomers(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
		ALTER TABLE customers
		DROP COLUMN IF EXISTS portal_password,
		DROP COLUMN IF EXISTS portal_last_login;
	`)
	return err
}
