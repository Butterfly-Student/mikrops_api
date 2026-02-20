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
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		payment_number VARCHAR(50) UNIQUE NOT NULL,

		customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE RESTRICT,
		invoice_id UUID REFERENCES invoices(id) ON DELETE RESTRICT, -- bisa NULL untuk advance payment

		-- Amount
		amount DECIMAL(12,2) NOT NULL,
		allocated_amount DECIMAL(12,2) DEFAULT 0, -- jumlah yang sudah dialokasikan ke invoice
		remaining_amount DECIMAL(12,2) GENERATED ALWAYS AS (amount - allocated_amount) STORED,

		-- Payment Details
		payment_method VARCHAR(20) NOT NULL CHECK (payment_method IN ('cash', 'bank_transfer', 'e-wallet', 'credit_card', 'debit_card', 'check', 'xendit')),
		payment_date TIMESTAMP NOT NULL,

		-- Bank Transfer Details
		bank_name VARCHAR(100),
		bank_account_number VARCHAR(50),
		bank_account_name VARCHAR(100),
		transaction_reference VARCHAR(100), -- nomor referensi/bukti transfer

		-- E-wallet Details
		ewallet_provider VARCHAR(50), -- gopay, ovo, dana, dll
		ewallet_number VARCHAR(50),

		-- Xendit Details
		xendit_invoice_id VARCHAR(100),
		xendit_external_id VARCHAR(100),
		xendit_payment_channel VARCHAR(50),

		-- Proof
		proof_image TEXT, -- URL/path foto bukti transfer
		receipt_number VARCHAR(50), -- nomor kwitansi

		-- Status
		status VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'confirmed', 'rejected', 'refunded')) DEFAULT 'pending',

		-- Processing
		processed_by UUID, -- FK to users yang konfirmasi
		processed_at TIMESTAMP,
		rejection_reason TEXT,

		-- Refund (jika ada)
		refund_amount DECIMAL(12,2) DEFAULT 0,
		refund_date TIMESTAMP,
		refund_reason TEXT,
		refunded_by UUID,

		notes TEXT,

		created_by UUID,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		deleted_at TIMESTAMP
	);

	-- Indexes
	CREATE INDEX IF NOT EXISTS idx_payments_customer ON payments(customer_id) WHERE deleted_at IS NULL;
	CREATE INDEX IF NOT EXISTS idx_payments_invoice ON payments(invoice_id) WHERE invoice_id IS NOT NULL AND deleted_at IS NULL;
	CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status) WHERE deleted_at IS NULL;
	CREATE INDEX IF NOT EXISTS idx_payments_date ON payments(payment_date) WHERE deleted_at IS NULL;
	CREATE INDEX IF NOT EXISTS idx_payments_method ON payments(payment_method) WHERE deleted_at IS NULL;
	CREATE INDEX IF NOT EXISTS idx_payments_number ON payments(payment_number);
	CREATE INDEX IF NOT EXISTS idx_payments_xendit_invoice ON payments(xendit_invoice_id) WHERE xendit_invoice_id IS NOT NULL;
	CREATE INDEX IF NOT EXISTS idx_payments_deleted ON payments(deleted_at);

	-- Comments
	COMMENT ON TABLE payments IS 'Pembayaran dari pelanggan';
	COMMENT ON COLUMN payments.remaining_amount IS 'Calculated: amount - allocated_amount, untuk advance payment';

	-- ============================================
	-- PAYMENT ALLOCATIONS TABLE
	-- ============================================
	CREATE TABLE IF NOT EXISTS payment_allocations (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		payment_id UUID NOT NULL REFERENCES payments(id) ON DELETE CASCADE,
		invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
		allocated_amount DECIMAL(12,2) NOT NULL,

		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,

		UNIQUE(payment_id, invoice_id)
	);

	-- Indexes
	CREATE INDEX IF NOT EXISTS idx_payment_allocations_payment ON payment_allocations(payment_id);
	CREATE INDEX IF NOT EXISTS idx_payment_allocations_invoice ON payment_allocations(invoice_id);

	-- Comments
	COMMENT ON TABLE payment_allocations IS 'Alokasi pembayaran ke invoice (untuk partial/multiple payment)';
	`)
	if err != nil {
		return err
	}
	return nil
}

func downPayments(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
	DROP TABLE IF EXISTS payment_allocations CASCADE;
	DROP TABLE IF EXISTS payments CASCADE;
	`)
	if err != nil {
		return err
	}
	return nil
}
