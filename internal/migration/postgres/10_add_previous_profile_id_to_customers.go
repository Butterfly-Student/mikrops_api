package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upAddPreviousProfileID, downAddPreviousProfileID)
}

func upAddPreviousProfileID(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	ALTER TABLE customers
	ADD COLUMN IF NOT EXISTS previous_profile_id UUID REFERENCES bandwidth_profiles(id) ON DELETE SET NULL;

	COMMENT ON COLUMN customers.previous_profile_id IS 'Stores original profile ID when customer is isolated, restored on un-isolation';
	`)
	return err
}

func downAddPreviousProfileID(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	ALTER TABLE customers DROP COLUMN IF EXISTS previous_profile_id;
	`)
	return err
}
