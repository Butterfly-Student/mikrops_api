package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upAddPortalFieldsToCustomers, downAddPortalFieldsToCustomers)
}

// upAddPortalFieldsToCustomers is a no-op — portal_password and portal_last_login
// were merged into the customers table in migration 2.
func upAddPortalFieldsToCustomers(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`SELECT 1; -- no-op: portal fields merged into customers (migration 2)`)
	return err
}

func downAddPortalFieldsToCustomers(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`SELECT 1; -- no-op`)
	return err
}
