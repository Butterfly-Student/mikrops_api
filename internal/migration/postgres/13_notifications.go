package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upNotifications, downNotifications)
}

func upNotifications(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS notifications (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			type VARCHAR(50) NOT NULL,
			recipient VARCHAR(255) NOT NULL,
			subject VARCHAR(500),
			content TEXT NOT NULL,
			customer_id UUID,
			invoice_id UUID,
			payment_id UUID,
			status VARCHAR(20) DEFAULT 'pending' NOT NULL,
			error_code VARCHAR(100),
			error_msg TEXT,
			retry_count INTEGER DEFAULT 0,
			scheduled_at TIMESTAMP,
			sent_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
			deleted_at TIMESTAMP,
			CONSTRAINT fk_notifications_customer FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE CASCADE,
			CONSTRAINT fk_notifications_invoice FOREIGN KEY (invoice_id) REFERENCES invoices(id) ON DELETE CASCADE,
			CONSTRAINT fk_notifications_payment FOREIGN KEY (payment_id) REFERENCES payments(id) ON DELETE CASCADE
		);
	`)

	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		CREATE INDEX IF NOT EXISTS idx_notifications_type ON notifications(type);
	`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		CREATE INDEX IF NOT EXISTS idx_notifications_status ON notifications(status);
	`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		CREATE INDEX IF NOT EXISTS idx_notifications_recipient ON notifications(recipient);
	`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		CREATE INDEX IF NOT EXISTS idx_notifications_customer_id ON notifications(customer_id);
	`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		CREATE INDEX IF NOT EXISTS idx_notifications_invoice_id ON notifications(invoice_id);
	`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		CREATE INDEX IF NOT EXISTS idx_notifications_payment_id ON notifications(payment_id);
	`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		CREATE INDEX IF NOT EXISTS idx_notifications_deleted_at ON notifications(deleted_at);
	`)
	if err != nil {
		return err
	}

	return nil
}

func downNotifications(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS notifications;`)
	return err
}
