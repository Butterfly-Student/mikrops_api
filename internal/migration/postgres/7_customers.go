package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upCustomerProfileFK, downCustomerProfileFK)
}

// upCustomerProfileFK adds the foreign key from customers.previous_profile_id
// to bandwidth_profiles, which is created in migration 6.
// The subscriptions table (migration 15) will hold PPPoE/service configuration.
func upCustomerProfileFK(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	-- Tambah FK previous_profile_id setelah bandwidth_profiles (migration 6) sudah ada
	ALTER TABLE customers
		ADD CONSTRAINT fk_customers_previous_profile
		FOREIGN KEY (previous_profile_id) REFERENCES bandwidth_profiles(id) ON DELETE SET NULL;

	-- Tambah FK created_by dan updated_by ke admin_users
	ALTER TABLE customers
		ADD CONSTRAINT fk_customers_created_by
		FOREIGN KEY (created_by) REFERENCES admin_users(id) ON DELETE SET NULL;

	ALTER TABLE customers
		ADD CONSTRAINT fk_customers_updated_by
		FOREIGN KEY (updated_by) REFERENCES admin_users(id) ON DELETE SET NULL;

	-- Index untuk status query yang sering dipakai
	CREATE INDEX IF NOT EXISTS idx_customers_previous_profile ON customers(previous_profile_id) WHERE previous_profile_id IS NOT NULL;

	COMMENT ON COLUMN customers.created_by IS 'Admin yang membuat data pelanggan';
	COMMENT ON COLUMN customers.updated_by IS 'Admin yang terakhir mengupdate data pelanggan';
	`)
	return err
}

func downCustomerProfileFK(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	ALTER TABLE customers DROP CONSTRAINT IF EXISTS fk_customers_previous_profile;
	ALTER TABLE customers DROP CONSTRAINT IF EXISTS fk_customers_created_by;
	ALTER TABLE customers DROP CONSTRAINT IF EXISTS fk_customers_updated_by;
	DROP INDEX IF EXISTS idx_customers_previous_profile;
	`)
	return err
}
