package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upCasbin, downCasbin)
}

func upCasbin(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`CREATE TABLE IF NOT EXISTS casbin_rule (
		id SERIAL PRIMARY KEY,
		ptype VARCHAR(100),
		v0 VARCHAR(100),
		v1 VARCHAR(100),
		v2 VARCHAR(100),
		v3 VARCHAR(100),
		v4 VARCHAR(100),
		v5 VARCHAR(100)
	);`)
	return err
}

func downCasbin(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE casbin_rule;`)
	return err
}
