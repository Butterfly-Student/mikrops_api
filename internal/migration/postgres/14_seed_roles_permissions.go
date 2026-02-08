package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upSeedRolesPermissions, downSeedRolesPermissions)
}

func upSeedRolesPermissions(ctx context.Context, tx *sql.Tx) error {
	// Seed Roles
	_, err := tx.Exec(`
		INSERT INTO roles (id, name, description) VALUES
		(1, 'admin', 'Administrator with full system access'),
		(2, 'technician', 'Technical staff for infrastructure management'),
		(3, 'sales', 'Sales team for customer management'),
		(4, 'finance', 'Finance team for payment and billing management')
		ON CONFLICT (name) DO NOTHING;
	`)
	if err != nil {
		return err
	}

	// Seed Permissions
	_, err = tx.Exec(`
		INSERT INTO permissions (name, resource, action, description) VALUES
		-- Tenant permissions
		('tenant_manage', 'tenant', 'manage', 'Full tenant management'),
		
		-- Staff permissions
		('staff_manage', 'staff', 'manage', 'Full staff management'),
		
		-- NAS permissions
		('nas_manage', 'nas', 'manage', 'Full NAS management'),
		('nas_read', 'nas', 'read', 'View NAS information'),
		
		-- Internet Package permissions
		('package_manage', 'internet_package', 'manage', 'Full internet package management'),
		('package_read', 'internet_package', 'read', 'View internet packages'),
		
		-- Customer permissions
		('customer_manage', 'customer', 'manage', 'Full customer management'),
		('customer_read', 'customer', 'read', 'View customer information'),
		
		-- Subscription permissions
		('subscription_manage', 'subscription', 'manage', 'Full subscription management'),
		('subscription_read', 'subscription', 'read', 'View subscription information'),
		
		-- Payment Method permissions
		('payment_method_manage', 'payment_method', 'manage', 'Full payment method management'),
		
		-- Invoice permissions
		('invoice_manage', 'invoice', 'manage', 'Full invoice management'),
		('invoice_read', 'invoice', 'read', 'View invoice information'),
		
		-- Payment permissions
		('payment_manage', 'payment', 'manage', 'Full payment management'),
		('payment_read', 'payment', 'read', 'View payment information'),
		
		-- Report permissions
		('report_manage', 'report', 'manage', 'Full report access')
		ON CONFLICT (name) DO NOTHING;
	`)
	if err != nil {
		return err
	}

	// Seed Role-Permission mappings
	_, err = tx.Exec(`
		INSERT INTO role_permissions (role_id, permission_id) 
		SELECT r.id, p.id FROM roles r, permissions p WHERE
		-- Admin gets all permissions
		(r.name = 'admin') OR
		
		-- Technician permissions
		(r.name = 'technician' AND p.name IN (
			'nas_manage', 'package_read', 'customer_read', 'subscription_read'
		)) OR
		
		-- Sales permissions
		(r.name = 'sales' AND p.name IN (
			'nas_read', 'package_read', 'customer_manage', 'subscription_manage', 'invoice_read', 'payment_read'
		)) OR
		
		-- Finance permissions
		(r.name = 'finance' AND p.name IN (
			'package_read', 'customer_read', 'subscription_read', 'payment_method_manage', 'invoice_manage', 'payment_manage', 'report_manage'
		))
		ON CONFLICT (role_id, permission_id) DO NOTHING;
	`)
	if err != nil {
		return err
	}

	return nil
}

func downSeedRolesPermissions(ctx context.Context, tx *sql.Tx) error {
	// Delete in reverse order of creation
	_, err := tx.Exec(`DELETE FROM role_permissions;`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`DELETE FROM permissions;`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`DELETE FROM roles;`)
	if err != nil {
		return err
	}

	return nil
}
