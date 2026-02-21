package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upCustomerRegistrations, downCustomerRegistrations)
}

func upCustomerRegistrations(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS customer_registrations (
			id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			full_name            VARCHAR(100) NOT NULL,
			email                VARCHAR(100),
			phone                VARCHAR(20)  NOT NULL,
			address              TEXT,
			latitude             DECIMAL(10,8),
			longitude            DECIMAL(11,8),
			notes                TEXT,
			bandwidth_profile_id UUID REFERENCES bandwidth_profiles(id) ON DELETE RESTRICT,
			preferred_router_id  UUID REFERENCES mikrotik_routers(id) ON DELETE SET NULL,
			ppp_secret_name      VARCHAR(100),
			status               VARCHAR(20)  NOT NULL DEFAULT 'pending',
			rejection_reason     TEXT,
			approved_by          BIGINT,
			approved_at          TIMESTAMPTZ,
			customer_id          UUID REFERENCES customers(id) ON DELETE SET NULL,
			created_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
			updated_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
			deleted_at           TIMESTAMPTZ
		);
		CREATE INDEX IF NOT EXISTS idx_customer_registrations_status     ON customer_registrations(status);
		CREATE INDEX IF NOT EXISTS idx_customer_registrations_deleted_at ON customer_registrations(deleted_at);
	`)
	return err
}

func downCustomerRegistrations(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `DROP TABLE IF EXISTS customer_registrations;`)
	return err
}
