package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upStaff, downStaff)
}

func upStaff(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`CREATE TABLE IF NOT EXISTS staffs (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		tenant_id UUID NOT NULL,
		role_id INTEGER NOT NULL,
		email VARCHAR(255) NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		full_name VARCHAR(255) NOT NULL,
		phone VARCHAR(50),
		is_active BOOLEAN DEFAULT true,
		last_login_at TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
		FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE RESTRICT
	);
	
	CREATE INDEX IF NOT EXISTS idx_staffs_tenant_id ON staffs(tenant_id);
	CREATE INDEX IF NOT EXISTS idx_staffs_role_id ON staffs(role_id);
	CREATE INDEX IF NOT EXISTS idx_staffs_email ON staffs(email);`)
	if err != nil {
		return err
	}
	return nil
}

func downStaff(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS staffs CASCADE;`)
	if err != nil {
		return err
	}
	return nil
}
