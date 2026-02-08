package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upPermission, downPermission)
}

func upPermission(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`CREATE TABLE IF NOT EXISTS permissions (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) UNIQUE NOT NULL,
		resource VARCHAR(100) NOT NULL,
		action VARCHAR(50) NOT NULL,
		description TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
	);
	
	CREATE INDEX IF NOT EXISTS idx_permissions_resource ON permissions(resource);
	CREATE INDEX IF NOT EXISTS idx_permissions_action ON permissions(action);`)
	if err != nil {
		return err
	}
	return nil
}

func downPermission(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS permissions CASCADE;`)
	if err != nil {
		return err
	}
	return nil
}
