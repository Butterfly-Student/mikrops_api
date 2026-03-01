package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upFixPaymentProcessedBy, downFixPaymentProcessedBy)
}

// upFixPaymentProcessedBy is a no-op — processed_by in payments is now UUID type
// with FK to admin_users (already set correctly in migration 9).
func upFixPaymentProcessedBy(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`SELECT 1; -- no-op: processed_by is UUID in migration 9`)
	return err
}

func downFixPaymentProcessedBy(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`SELECT 1; -- no-op`)
	return err
}
