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
	-- ============================================
	-- MESSAGE TEMPLATES TABLE
	-- ============================================
	CREATE TABLE IF NOT EXISTS message_templates (
		id         UUID     PRIMARY KEY DEFAULT gen_random_uuid(),
		event      VARCHAR(80) UNIQUE NOT NULL, -- mis: 'invoice_created', 'payment_confirmed', 'subscription_suspended'
		channel    VARCHAR(20) CHECK (channel IN ('whatsapp', 'email', 'both')) DEFAULT 'whatsapp',
		subject    VARCHAR(200),   -- khusus email, NULL untuk WA
		body       TEXT NOT NULL,  -- template dengan variabel {{nama}}, {{invoice_no}}, dll
		is_active  BOOLEAN DEFAULT true,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_message_templates_event     ON message_templates(event);
	CREATE INDEX IF NOT EXISTS idx_message_templates_channel   ON message_templates(channel);
	CREATE INDEX IF NOT EXISTS idx_message_templates_is_active ON message_templates(is_active);

	COMMENT ON TABLE  message_templates         IS 'Template pesan WA/Email per event sistem';
	COMMENT ON COLUMN message_templates.event   IS 'Nama event, mis: invoice_created, payment_confirmed';
	COMMENT ON COLUMN message_templates.body    IS 'Template dengan variabel Go template: {{.Nama}}, {{.InvoiceNo}}, dll';
	COMMENT ON COLUMN message_templates.subject IS 'Khusus email subject, NULL untuk WA';
	`)
	return err
}

func downMessageTemplates(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS message_templates CASCADE;`)
	return err
}
