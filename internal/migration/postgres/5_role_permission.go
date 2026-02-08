package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upRolePermission, downRolePermission)
}

func upRolePermission(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`CREATE TABLE IF NOT EXISTS role_permissions (
		role_id INTEGER NOT NULL,
		permission_id INTEGER NOT NULL,
		PRIMARY KEY (role_id, permission_id),
		FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
		FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
	);`)
	if err != nil {
		return err
	}
	return nil
}

func downRolePermission(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS role_permissions CASCADE;`)
	if err != nil {
		return err
	}
	return nil
}
