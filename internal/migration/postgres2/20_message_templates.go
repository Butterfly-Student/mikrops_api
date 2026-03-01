package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upMessageTemplates, downMessageTemplates)
}

func upMessageTemplates(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	CREATE TABLE IF NOT EXISTS message_templates (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		event VARCHAR(80) UNIQUE NOT NULL,
		channel VARCHAR(20) CHECK (channel IN ('whatsapp', 'email', 'both')) DEFAULT 'whatsapp',
		subject VARCHAR(200),
		body TEXT NOT NULL,
		is_active BOOLEAN DEFAULT true,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_message_templates_event ON message_templates(event);
	CREATE INDEX IF NOT EXISTS idx_message_templates_channel ON message_templates(channel);
	CREATE INDEX IF NOT EXISTS idx_message_templates_is_active ON message_templates(is_active);

	COMMENT ON TABLE message_templates IS 'Template pesan WA dan Email per event';
	COMMENT ON COLUMN message_templates.subject IS 'khusus email, NULL untuk WA';
	`)
	return err
}

func downMessageTemplates(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS message_templates CASCADE;`)
	return err
}
