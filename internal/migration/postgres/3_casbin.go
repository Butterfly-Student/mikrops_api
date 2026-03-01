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
	_, err := tx.Exec(`
	CREATE TABLE IF NOT EXISTS casbin_rule (
		id    BIGSERIAL   PRIMARY KEY,
		ptype VARCHAR(10),
		v0    VARCHAR(256),
		v1    VARCHAR(256),
		v2    VARCHAR(256),
		v3    VARCHAR(256),
		v4    VARCHAR(256),
		v5    VARCHAR(256)
	);

	CREATE INDEX IF NOT EXISTS idx_casbin_rule_ptype ON casbin_rule(ptype);
	CREATE INDEX IF NOT EXISTS idx_casbin_rule_v0    ON casbin_rule(v0);
	CREATE INDEX IF NOT EXISTS idx_casbin_rule_v1    ON casbin_rule(v1);

	COMMENT ON TABLE casbin_rule IS 'Casbin RBAC rules untuk otorisasi berbasis peran';
	`)
	return err
}

func downCasbin(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS casbin_rule CASCADE;`)
	return err
}
