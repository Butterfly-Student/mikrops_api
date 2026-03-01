package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upAdminUsers, downAdminUsers)
}

func upAdminUsers(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	CREATE TABLE IF NOT EXISTS admin_users (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		full_name VARCHAR(100) NOT NULL,
		email VARCHAR(100) UNIQUE NOT NULL,
		phone VARCHAR(20),
		password_hash VARCHAR(255) NOT NULL,
		role VARCHAR(20) NOT NULL CHECK (role IN ('superadmin', 'admin', 'cs', 'billing', 'technician', 'readonly')) DEFAULT 'cs',
		is_active BOOLEAN DEFAULT true,
		last_login TIMESTAMP,
		last_ip VARCHAR(45),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_admin_users_email ON admin_users(email);
	CREATE INDEX IF NOT EXISTS idx_admin_users_role ON admin_users(role);
	CREATE INDEX IF NOT EXISTS idx_admin_users_is_active ON admin_users(is_active);

	COMMENT ON TABLE admin_users IS 'Admin dan operator sistem';
	COMMENT ON COLUMN admin_users.password_hash IS 'bcrypt hash';
	`)
	return err
}

func downAdminUsers(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS admin_users CASCADE;`)
	return err
}
