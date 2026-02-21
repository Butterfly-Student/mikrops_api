package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upFixPaymentProcessedBy, downFixPaymentProcessedBy)
}

func upFixPaymentProcessedBy(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	ALTER TABLE payments
	ALTER COLUMN processed_by TYPE BIGINT USING NULL;
	`)
	return err
}

func downFixPaymentProcessedBy(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	ALTER TABLE payments
	ALTER COLUMN processed_by TYPE UUID USING NULL;
	`)
	return err
}
