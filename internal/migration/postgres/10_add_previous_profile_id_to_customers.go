package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upAddPreviousProfileID, downAddPreviousProfileID)
}

// upAddPreviousProfileID is a no-op — previous_profile_id was merged into 2_user.go.
// Kept to preserve goose migration history / sequence.
func upAddPreviousProfileID(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	-- No-op: previous_profile_id already included in customers table (migration 2)
	-- and FK constraint added in migration 7.
	SELECT 1;
	`)
	return err
}

func downAddPreviousProfileID(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`SELECT 1; -- no-op`)
	return err
}
