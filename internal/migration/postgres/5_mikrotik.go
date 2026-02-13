package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upMikrotik, downMikrotik)
}

func upMikrotik(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`CREATE TABLE IF NOT EXISTS mikrotik_routers (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		address VARCHAR(100) NOT NULL,
		username VARCHAR(100) NOT NULL,
		password VARCHAR(255) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		deleted_at TIMESTAMP
	);`)
	return err
}

func downMikrotik(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE mikrotik_routers;`)
	return err
}
