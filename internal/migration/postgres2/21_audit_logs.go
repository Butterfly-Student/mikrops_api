package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upAuditLogs, downAuditLogs)
}

func upAuditLogs(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	CREATE TABLE IF NOT EXISTS audit_logs (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		admin_id UUID REFERENCES admin_users(id) ON DELETE SET NULL,
		action VARCHAR(100) NOT NULL,
		entity_type VARCHAR(50) NOT NULL,
		entity_id UUID NOT NULL,
		old_value JSONB,
		new_value JSONB,
		ip_address VARCHAR(45),
		user_agent VARCHAR(255),
		notes TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_audit_logs_admin ON audit_logs(admin_id);
	CREATE INDEX IF NOT EXISTS idx_audit_logs_entity ON audit_logs(entity_type, entity_id);
	CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
	CREATE INDEX IF NOT EXISTS idx_audit_logs_created ON audit_logs(created_at DESC);
	CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);

	COMMENT ON TABLE audit_logs IS 'Audit trail semua perubahan penting di sistem';
	COMMENT ON COLUMN audit_logs.admin_id IS 'NULL = dilakukan oleh sistem/scheduler';
	`)
	return err
}

func downAuditLogs(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS audit_logs CASCADE;`)
	return err
}
