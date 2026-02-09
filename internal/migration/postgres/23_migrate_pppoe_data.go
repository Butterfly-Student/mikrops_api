package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upMigratePppoeData, downMigratePppoeData)
}

func upMigratePppoeData(ctx context.Context, tx *sql.Tx) error {
	// Migrate existing PPPoE data from customers to pppoe_accounts
	// Only migrate customers that have pppoe_username set
	_, err := tx.Exec(`INSERT INTO pppoe_accounts (tenant_id, customer_id, nas_id, package_id, username, password_encrypted, status, sync_status)
	SELECT
		c.tenant_id,
		c.id,
		c.nas_id,
		COALESCE(
			(SELECT s.package_id FROM subscriptions s WHERE s.customer_id = c.id AND s.status = 'active' LIMIT 1),
			(SELECT ip.id FROM internet_packages ip WHERE ip.tenant_id = c.tenant_id LIMIT 1)
		),
		c.pppoe_username,
		c.pppoe_password,
		CASE WHEN c.is_active THEN 'active' ELSE 'disabled' END,
		'synced'
	FROM customers c
	WHERE c.pppoe_username IS NOT NULL
		AND c.pppoe_username != ''
		AND c.nas_id IS NOT NULL;`)
	if err != nil {
		return err
	}
	return nil
}

func downMigratePppoeData(ctx context.Context, tx *sql.Tx) error {
	// Remove migrated data
	_, err := tx.Exec(`DELETE FROM pppoe_accounts WHERE sync_status = 'synced';`)
	if err != nil {
		return err
	}
	return nil
}
