package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upSystemSettings, downSystemSettings)
}

func upSystemSettings(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	-- ============================================
	-- SYSTEM SETTINGS TABLE
	-- ============================================
	CREATE TABLE IF NOT EXISTS system_settings (
		id           UUID     PRIMARY KEY DEFAULT gen_random_uuid(),
		group_name   VARCHAR(50) NOT NULL,
		key_name     VARCHAR(100) NOT NULL,
		value        TEXT,
		type         VARCHAR(20) CHECK (type IN ('string', 'integer', 'boolean', 'json', 'password')) DEFAULT 'string',
		label        VARCHAR(150),
		description  TEXT,
		is_encrypted BOOLEAN DEFAULT false,
		is_public    BOOLEAN DEFAULT false, -- bisa dibaca frontend tanpa auth
		updated_at   TIMESTAMP,
		updated_by   UUID REFERENCES admin_users(id) ON DELETE SET NULL,
		UNIQUE (group_name, key_name)
	);

	CREATE INDEX IF NOT EXISTS idx_system_settings_group ON system_settings(group_name);
	CREATE INDEX IF NOT EXISTS idx_system_settings_public ON system_settings(is_public) WHERE is_public = true;

	COMMENT ON TABLE  system_settings              IS 'Konfigurasi sistem dalam format key-value per group';
	COMMENT ON COLUMN system_settings.is_encrypted IS 'jika TRUE, value disimpan terenkripsi AES-256';
	COMMENT ON COLUMN system_settings.is_public    IS 'jika TRUE, bisa dibaca tanpa auth (untuk frontend config)';

	-- ============================================
	-- SETTINGS HISTORY TABLE (audit trail settings)
	-- ============================================
	CREATE TABLE IF NOT EXISTS settings_history (
		id         UUID     PRIMARY KEY DEFAULT gen_random_uuid(),
		group_name VARCHAR(50)  NOT NULL,
		key_name   VARCHAR(100) NOT NULL,
		old_value  TEXT,
		new_value  TEXT,
		changed_by UUID     NOT NULL REFERENCES admin_users(id) ON DELETE RESTRICT,
		changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		ip_address VARCHAR(45),
		user_agent VARCHAR(255)
	);

	CREATE INDEX IF NOT EXISTS idx_settings_history_setting ON settings_history(group_name, key_name);
	CREATE INDEX IF NOT EXISTS idx_settings_history_changed ON settings_history(changed_at DESC);
	CREATE INDEX IF NOT EXISTS idx_settings_history_admin   ON settings_history(changed_by);

	COMMENT ON TABLE settings_history IS 'Riwayat perubahan system_settings';
	`)
	return err
}

func downSystemSettings(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	DROP TABLE IF EXISTS settings_history CASCADE;
	DROP TABLE IF EXISTS system_settings  CASCADE;
	`)
	return err
}
