package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upSettingsHistory, downSettingsHistory)
}

func upSettingsHistory(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	CREATE TABLE IF NOT EXISTS settings_history (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		group_name VARCHAR(50) NOT NULL,
		key_name VARCHAR(100) NOT NULL,
		old_value TEXT,
		new_value TEXT,
		changed_by UUID NOT NULL REFERENCES admin_users(id),
		changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		ip_address VARCHAR(45),
		user_agent VARCHAR(255)
	);

	CREATE INDEX IF NOT EXISTS idx_settings_history_setting ON settings_history(group_name, key_name);
	CREATE INDEX IF NOT EXISTS idx_settings_history_changed ON settings_history(changed_at);
	CREATE INDEX IF NOT EXISTS idx_settings_history_admin ON settings_history(changed_by);

	COMMENT ON TABLE settings_history IS 'Riwayat perubahan system settings';
	`)
	return err
}

func downSettingsHistory(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS settings_history CASCADE;`)
	return err
}
