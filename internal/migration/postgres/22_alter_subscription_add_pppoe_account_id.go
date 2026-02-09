package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upAlterSubscriptionAddPppoeAccountID, downAlterSubscriptionAddPppoeAccountID)
}

func upAlterSubscriptionAddPppoeAccountID(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS pppoe_account_id UUID;
	ALTER TABLE subscriptions ADD CONSTRAINT fk_subscriptions_pppoe_account FOREIGN KEY (pppoe_account_id) REFERENCES pppoe_accounts(id) ON DELETE SET NULL;
	CREATE INDEX IF NOT EXISTS idx_subscriptions_pppoe_account_id ON subscriptions(pppoe_account_id);`)
	if err != nil {
		return err
	}
	return nil
}

func downAlterSubscriptionAddPppoeAccountID(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`ALTER TABLE subscriptions DROP CONSTRAINT IF EXISTS fk_subscriptions_pppoe_account;
	ALTER TABLE subscriptions DROP COLUMN IF EXISTS pppoe_account_id;`)
	if err != nil {
		return err
	}
	return nil
}
