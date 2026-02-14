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
	_, err := tx.Exec(`
	-- Create ENUM for setting value type
	DO $$ BEGIN
		CREATE TYPE setting_value_type AS ENUM ('string', 'number', 'boolean', 'json');
	EXCEPTION
		WHEN duplicate_object THEN null;
	END $$;

	CREATE TABLE IF NOT EXISTS system_settings (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		key VARCHAR(255) UNIQUE NOT NULL,
		value TEXT,
		value_type setting_value_type NOT NULL DEFAULT 'string',
		category VARCHAR(100) NOT NULL,
		description TEXT,
		is_public BOOLEAN DEFAULT false,
		updated_by UUID REFERENCES users(id) ON DELETE SET NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_system_settings_key ON system_settings(key);
	CREATE INDEX IF NOT EXISTS idx_system_settings_category ON system_settings(category);
	CREATE INDEX IF NOT EXISTS idx_system_settings_is_public ON system_settings(is_public);

	-- Insert default settings
	INSERT INTO system_settings (key, value, value_type, category, description, is_public) VALUES
	-- Billing settings
	('invoice.due_days', '7', 'number', 'billing', 'Number of days until invoice is due', false),
	('invoice.grace_period_days', '3', 'number', 'billing', 'Grace period before customer isolation', false),
	('invoice.late_fee_enabled', 'true', 'boolean', 'billing', 'Enable late fee charges', false),
	('invoice.late_fee_amount', '50000', 'number', 'billing', 'Late fee amount in IDR', false),
	('invoice.auto_generate_day', '25', 'number', 'billing', 'Day of month to auto-generate invoices', false),
	('invoice.tax_rate', '0.11', 'number', 'billing', 'Default tax rate (PPN 11%)', false),

	-- Company information
	('company.name', 'PT Internet Provider', 'string', 'company', 'Company name', true),
	('company.address', 'Jl. Raya No. 123', 'string', 'company', 'Company address', true),
	('company.phone', '021-12345678', 'string', 'company', 'Company phone number', true),
	('company.email', 'info@isp.com', 'string', 'company', 'Company email', true),
	('company.tax_id', '01.234.567.8-901.000', 'string', 'company', 'Company tax ID (NPWP)', true),
	('company.logo_url', '', 'string', 'company', 'Company logo URL', true),

	-- WhatsApp settings
	('whatsapp.enabled', 'true', 'boolean', 'whatsapp', 'Enable WhatsApp notifications', false),
	('whatsapp.group_billing', '', 'string', 'whatsapp', 'WhatsApp group ID for billing notifications', false),
	('whatsapp.group_support', '', 'string', 'whatsapp', 'WhatsApp group ID for support notifications', false),
	('whatsapp.group_notifications', '', 'string', 'whatsapp', 'WhatsApp group ID for general notifications', false),
	('whatsapp.api_url', 'http://localhost:3000', 'string', 'whatsapp', 'Gowa WhatsApp API URL', false),
	('whatsapp.api_key', '', 'string', 'whatsapp', 'Gowa WhatsApp API key', false),

	-- Xendit settings
	('xendit.secret_key', '', 'string', 'xendit', 'Xendit secret key', false),
	('xendit.api_key', '', 'string', 'xendit', 'Xendit API key', false),
	('xendit.webhook_token', '', 'string', 'xendit', 'Xendit webhook verification token', false),
	('xendit.environment', 'development', 'string', 'xendit', 'Xendit environment (development/production)', false),

	-- Payment portal settings
	('payment_portal.enabled', 'true', 'boolean', 'payment', 'Enable payment portal', true),
	('payment_portal.url', 'https://portal.example.com', 'string', 'payment', 'Payment portal URL', true),
	('payment_portal.allow_methods', '["va","e-wallet","credit_card","qris"]', 'json', 'payment', 'Allowed payment methods', false),

	-- System settings
	('system.maintenance_mode', 'false', 'boolean', 'system', 'Enable maintenance mode', false),
	('system.maintenance_message', 'System is under maintenance', 'string', 'system', 'Maintenance mode message', true),
	('system.timezone', 'Asia/Jakarta', 'string', 'system', 'System timezone', false),
	('system.date_format', 'DD/MM/YYYY', 'string', 'system', 'Date format', false),
	('system.currency', 'IDR', 'string', 'system', 'Currency code', true)
	ON CONFLICT (key) DO NOTHING;
	`)
	if err != nil {
		return err
	}
	return nil
}

func downSystemSetting(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	DROP TABLE IF EXISTS system_settings CASCADE;
	DROP TYPE IF EXISTS setting_value_type CASCADE;
	`)
	if err != nil {
		return err
	}
	return nil
}
