package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upAlterInternetPackageAddProfileName, downAlterInternetPackageAddProfileName)
}

func upAlterInternetPackageAddProfileName(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`ALTER TABLE internet_packages ADD COLUMN IF NOT EXISTS profile_name VARCHAR(100);`)
	if err != nil {
		return err
	}
	return nil
}

func downAlterInternetPackageAddProfileName(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`ALTER TABLE internet_packages DROP COLUMN IF EXISTS profile_name;`)
	if err != nil {
		return err
	}
	return nil
}
