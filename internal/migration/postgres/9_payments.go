package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upPayments, downPayments)
}

func upPayments(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	-- ============================================
	-- PAYMENTS TABLE
	-- ============================================
	CREATE TABLE IF NOT EXISTS payments (
		id             UUID      PRIMARY KEY DEFAULT gen_random_uuid(),
		payment_number VARCHAR(50) UNIQUE NOT NULL,

		customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE RESTRICT,
		invoice_id  UUID REFERENCES invoices(id) ON DELETE RESTRICT, -- NULL untuk advance payment

		-- Nominal
		amount           DECIMAL(12,2) NOT NULL,
		allocated_amount DECIMAL(12,2) DEFAULT 0,
		remaining_amount DECIMAL(12,2) GENERATED ALWAYS AS (amount - allocated_amount) STORED,

		-- Metode Pembayaran
		payment_method VARCHAR(20) NOT NULL
		               CHECK (payment_method IN ('cash', 'bank_transfer', 'e-wallet', 'credit_card', 'debit_card', 'check', 'qris', 'gateway')),
		payment_date   TIMESTAMP  NOT NULL,

		-- Detail Transfer Bank
		bank_name           VARCHAR(100),
		bank_account_number VARCHAR(50),
		bank_account_name   VARCHAR(100),
		transaction_reference VARCHAR(100),

		-- Detail E-Wallet
		ewallet_provider VARCHAR(50),
		ewallet_number   VARCHAR(50),

		-- Detail Payment Gateway (Xendit, Midtrans, dll)
		gateway_name     VARCHAR(50),   -- nama gateway: xendit, midtrans, dll
		gateway_trx_id   VARCHAR(150),  -- ID transaksi dari gateway
		gateway_response JSONB,         -- raw response JSON dari gateway

		-- Backward-compat Xendit fields
		xendit_invoice_id      VARCHAR(100),
		xendit_external_id     VARCHAR(100),
		xendit_payment_channel VARCHAR(50),

		-- Bukti Bayar
		proof_image    TEXT,   -- URL foto bukti transfer
		receipt_number VARCHAR(50),

		-- Status
		status VARCHAR(20) NOT NULL
		       CHECK (status IN ('pending', 'confirmed', 'rejected', 'refunded'))
		       DEFAULT 'pending',

		-- Verifikasi
		processed_by UUID REFERENCES admin_users(id) ON DELETE SET NULL,
		processed_at TIMESTAMP,
		rejection_reason TEXT,

		-- Refund
		refund_amount DECIMAL(12,2) DEFAULT 0,
		refund_date   TIMESTAMP,
		refund_reason TEXT,
		refunded_by   UUID,

		notes      TEXT,
		created_by UUID,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		deleted_at TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_payments_customer       ON payments(customer_id)    WHERE deleted_at IS NULL;
	CREATE INDEX IF NOT EXISTS idx_payments_invoice        ON payments(invoice_id)     WHERE invoice_id IS NOT NULL AND deleted_at IS NULL;
	CREATE INDEX IF NOT EXISTS idx_payments_status         ON payments(status)         WHERE deleted_at IS NULL;
	CREATE INDEX IF NOT EXISTS idx_payments_date           ON payments(payment_date)   WHERE deleted_at IS NULL;
	CREATE INDEX IF NOT EXISTS idx_payments_method         ON payments(payment_method) WHERE deleted_at IS NULL;
	CREATE INDEX IF NOT EXISTS idx_payments_number         ON payments(payment_number);
	CREATE INDEX IF NOT EXISTS idx_payments_gateway        ON payments(gateway_name, gateway_trx_id) WHERE gateway_trx_id IS NOT NULL;
	CREATE INDEX IF NOT EXISTS idx_payments_xendit_invoice ON payments(xendit_invoice_id) WHERE xendit_invoice_id IS NOT NULL;
	CREATE INDEX IF NOT EXISTS idx_payments_deleted        ON payments(deleted_at);

	COMMENT ON TABLE  payments                IS 'Pembayaran dari pelanggan';
	COMMENT ON COLUMN payments.remaining_amount IS 'Calculated: amount - allocated_amount (untuk advance payment)';
	COMMENT ON COLUMN payments.gateway_name    IS 'Nama gateway: xendit, midtrans, dll';
	COMMENT ON COLUMN payments.gateway_response IS 'Raw response JSON dari payment gateway';

	-- ============================================
	-- PAYMENT ALLOCATIONS TABLE
	-- ============================================
	CREATE TABLE IF NOT EXISTS payment_allocations (
		id               UUID      PRIMARY KEY DEFAULT gen_random_uuid(),
		payment_id       UUID      NOT NULL REFERENCES payments(id) ON DELETE CASCADE,
		invoice_id       UUID      NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
		allocated_amount DECIMAL(12,2) NOT NULL,
		created_at       TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		UNIQUE(payment_id, invoice_id)
	);

	CREATE INDEX IF NOT EXISTS idx_payment_allocations_payment ON payment_allocations(payment_id);
	CREATE INDEX IF NOT EXISTS idx_payment_allocations_invoice ON payment_allocations(invoice_id);

	COMMENT ON TABLE payment_allocations IS 'Alokasi 1 pembayaran ke multiple invoice (partial/advance payment)';
	`)
	return err
}

func downPayments(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	DROP TABLE IF EXISTS payment_allocations CASCADE;
	DROP TABLE IF EXISTS payments            CASCADE;
	`)
	return err
}
