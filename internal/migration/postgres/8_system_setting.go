package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upSystemSetting, downSystemSetting)
}

func upSystemSetting(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`CREATE TABLE system_settings (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		key VARCHAR(100) UNIQUE NOT NULL,
		value TEXT,
		value_type VARCHAR(20) NOT NULL DEFAULT 'string' CHECK (value_type IN ('string', 'number', 'boolean', 'json')),
		category VARCHAR(50),
		description TEXT,
		is_public BOOLEAN DEFAULT false,
		updated_by INTEGER REFERENCES users(id),
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
	);`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`INSERT INTO system_settings (key, value, value_type, category, description) VALUES
	-- Invoice Settings
	('invoice.due_days', '7', 'number', 'invoice', 'Jumlah hari dari issue date ke due date'),
	('invoice.grace_period_days', '3', 'number', 'invoice', 'Grace period sebelum auto isolate'),
	('invoice.late_fee_enabled', 'true', 'boolean', 'invoice', 'Enable late fee'),
	('invoice.late_fee_amount', '50000', 'number', 'invoice', 'Denda keterlambatan (Rp)'),
	('invoice.auto_generate_day', '25', 'number', 'invoice', 'Tanggal generate invoice otomatis'),

	-- Company Settings
	('company.name', 'PT Internet Provider', 'string', 'company', 'Nama perusahaan'),
	('company.address', 'Jl. Raya No. 123', 'string', 'company', 'Alamat perusahaan'),
	('company.phone', '021-12345678', 'string', 'company', 'Telepon perusahaan'),
	('company.email', 'info@isp.com', 'string', 'company', 'Email perusahaan'),
	('company.tax_id', '01.234.567.8-901.000', 'string', 'company', 'NPWP'),

	-- Email Settings
	('email.smtp_host', 'smtp.gmail.com', 'string', 'email', 'SMTP Host'),
	('email.smtp_port', '587', 'number', 'email', 'SMTP Port'),
	('email.smtp_user', '', 'string', 'email', 'SMTP Username'),
	('email.smtp_password', '', 'string', 'email', 'SMTP Password'),
	('email.from_email', 'noreply@example.com', 'string', 'email', 'From Email Address'),
	('email.from_name', 'Your ISP', 'string', 'email', 'From Name'),
	('email.enabled', 'false', 'boolean', 'email', 'Enable Email Notifications'),

	-- Gowa WhatsApp Settings
	('gowa.api_url', 'http://localhost:3000', 'string', 'gowa', 'Gowa API URL'),
	('gowa.api_key', '', 'string', 'gowa', 'Gowa API Key'),
	('gowa.enabled', 'false', 'boolean', 'gowa', 'Enable Gowa WhatsApp Notifications'),
	('gowa.timeout', '30', 'number', 'gowa', 'Gowa Request Timeout (seconds)'),

	-- WhatsApp Group IDs
	('whatsapp.group_billing', '', 'string', 'whatsapp', 'WhatsApp Group ID untuk notifikasi billing'),
	('whatsapp.group_support', '', 'string', 'whatsapp', 'WhatsApp Group ID untuk support'),
	('whatsapp.group_notifications', '', 'string', 'whatsapp', 'WhatsApp Group ID untuk notifikasi PPPoE'),

	-- Xendit Payment Settings
	('xendit.secret_key', '', 'string', 'xendit', 'Xendit Secret Key'),
	('xendit.api_key', '', 'string', 'xendit', 'Xendit API Key'),
	('xendit.environment', 'development', 'string', 'xendit', 'Xendit Environment (development/production)'),

	-- Payment Portal Settings
	('payment.portal_url', 'https://portal.example.com', 'string', 'payment', 'URL Payment Portal'),
	('payment.portal_enabled', 'true', 'boolean', 'payment', 'Enable Payment Portal');`)
	if err != nil {
		return err
	}

	return nil
}

func downSystemSetting(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS system_settings CASCADE;`)
	return err
}
