package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PaymentMethod represents the method used for payment
type PaymentMethod string

const (
	PaymentMethodCash         PaymentMethod = "cash"
	PaymentMethodBankTransfer PaymentMethod = "bank_transfer"
	PaymentMethodEWallet      PaymentMethod = "e-wallet"
	PaymentMethodCreditCard   PaymentMethod = "credit_card"
	PaymentMethodDebitCard    PaymentMethod = "debit_card"
	PaymentMethodCheck        PaymentMethod = "check"
	PaymentMethodQRIS         PaymentMethod = "qris"
	PaymentMethodXendit       PaymentMethod = "xendit"
	PaymentMethodGateway      PaymentMethod = "gateway" // generic gateway
)

// PaymentStatusType represents the confirmation state of a payment record
// (PaymentStatus type — invoice coverage — is declared in invoice.go)
type PaymentStatusType string

const (
	PaymentStatusTypePending   PaymentStatusType = "pending"
	PaymentStatusTypeConfirmed PaymentStatusType = "confirmed"
	PaymentStatusTypeRejected  PaymentStatusType = "rejected"
	PaymentStatusTypeRefunded  PaymentStatusType = "refunded"
)

// Payment represents a payment transaction made by a customer
type Payment struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PaymentNumber string    `gorm:"size:50;uniqueIndex;not null" json:"payment_number" validate:"required"`

	CustomerID uuid.UUID `gorm:"type:uuid;not null" json:"customer_id" validate:"required"`
	Customer   *Customer `gorm:"foreignKey:CustomerID;constraint:OnDelete:RESTRICT" json:"customer,omitempty"`

	InvoiceID *uuid.UUID `gorm:"type:uuid" json:"invoice_id,omitempty"` // nullable for advance payment
	Invoice   *Invoice   `gorm:"foreignKey:InvoiceID;constraint:OnDelete:RESTRICT" json:"invoice,omitempty"`

	// Amount
	Amount          float64 `gorm:"type:decimal(12,2);not null" json:"amount" validate:"required,min=0"`
	AllocatedAmount float64 `gorm:"type:decimal(12,2);default:0" json:"allocated_amount"`
	RemainingAmount float64 `gorm:"type:decimal(12,2);->;default:0" json:"remaining_amount"` // computed

	// Payment details
	PaymentMethod PaymentMethod `gorm:"size:20;not null" json:"payment_method" validate:"required"`
	PaymentDate   time.Time     `gorm:"not null" json:"payment_date" validate:"required"`

	// Bank Transfer
	BankName             *string `gorm:"size:100" json:"bank_name,omitempty"`
	BankAccountNumber    *string `gorm:"size:50" json:"bank_account_number,omitempty"`
	BankAccountName      *string `gorm:"size:100" json:"bank_account_name,omitempty"`
	TransactionReference *string `gorm:"size:100" json:"transaction_reference,omitempty"`

	// E-Wallet
	EwalletProvider *string `gorm:"size:50" json:"ewallet_provider,omitempty"`
	EwalletNumber   *string `gorm:"size:50" json:"ewallet_number,omitempty"`

	// ─── Generic Payment Gateway ────────────────────────────────
	GatewayName     *string                `gorm:"size:50" json:"gateway_name,omitempty"`                        // xendit, midtrans, etc.
	GatewayTrxID    *string                `gorm:"size:100" json:"gateway_trx_id,omitempty"`                     // transaction ID from gateway
	GatewayResponse map[string]interface{} `gorm:"serializer:json;type:jsonb" json:"gateway_response,omitempty"` // full gateway response

	// Legacy Xendit fields (kept for backward compat)
	XenditInvoiceID      *string `gorm:"size:100" json:"xendit_invoice_id,omitempty"`
	XenditExternalID     *string `gorm:"size:100" json:"xendit_external_id,omitempty"`
	XenditPaymentChannel *string `gorm:"size:50" json:"xendit_payment_channel,omitempty"`

	// Proof
	ProofImage    *string `gorm:"type:text" json:"proof_image,omitempty"`
	ReceiptNumber *string `gorm:"size:50" json:"receipt_number,omitempty"`

	// Status
	Status PaymentStatusType `gorm:"size:20;not null;default:pending" json:"status"`

	// Processing
	ProcessedBy     *uuid.UUID `gorm:"type:uuid" json:"processed_by,omitempty"` // FK → admin_users (changed from *uint)
	ProcessedAt     *time.Time `json:"processed_at,omitempty"`
	RejectionReason *string    `gorm:"type:text" json:"rejection_reason,omitempty"`

	// Refund
	RefundAmount *float64   `gorm:"type:decimal(12,2);default:0" json:"refund_amount,omitempty"`
	RefundDate   *time.Time `json:"refund_date,omitempty"`
	RefundReason *string    `gorm:"type:text" json:"refund_reason,omitempty"`
	RefundedBy   *uuid.UUID `gorm:"type:uuid" json:"refunded_by,omitempty"` // FK → admin_users (changed from *uint)

	Notes *string `gorm:"type:text" json:"notes,omitempty"`

	// Relations
	Allocations []PaymentAllocation `gorm:"foreignKey:PaymentID" json:"allocations,omitempty"`

	CreatedBy *uuid.UUID     `gorm:"type:uuid" json:"created_by,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name
func (Payment) TableName() string { return "payments" }

// PaymentAllocation represents an allocation of a payment to an invoice
type PaymentAllocation struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PaymentID       uuid.UUID `gorm:"type:uuid;not null" json:"payment_id" validate:"required"`
	Payment         *Payment  `gorm:"foreignKey:PaymentID;constraint:OnDelete:CASCADE" json:"payment,omitempty"`
	InvoiceID       uuid.UUID `gorm:"type:uuid;not null" json:"invoice_id" validate:"required"`
	Invoice         *Invoice  `gorm:"foreignKey:InvoiceID;constraint:OnDelete:CASCADE" json:"invoice,omitempty"`
	AllocatedAmount float64   `gorm:"type:decimal(12,2);not null" json:"allocated_amount" validate:"required,min=0"`
	CreatedAt       time.Time `json:"created_at"`
}

func (PaymentAllocation) TableName() string { return "payment_allocations" }

// PaymentInput for creating/updating a payment
type PaymentInput struct {
	PaymentNumber        string     `json:"payment_number" validate:"required,max=50"`
	CustomerID           uuid.UUID  `json:"customer_id" validate:"required"`
	InvoiceID            *uuid.UUID `json:"invoice_id" validate:"omitempty"`
	Amount               float64    `json:"amount" validate:"required,min=0"`
	PaymentMethod        string     `json:"payment_method" validate:"required,oneof=cash bank_transfer e-wallet credit_card debit_card check qris xendit gateway"`
	PaymentDate          time.Time  `json:"payment_date" validate:"required"`
	BankName             *string    `json:"bank_name" validate:"omitempty,max=100"`
	BankAccountNumber    *string    `json:"bank_account_number" validate:"omitempty,max=50"`
	BankAccountName      *string    `json:"bank_account_name" validate:"omitempty,max=100"`
	TransactionReference *string    `json:"transaction_reference" validate:"omitempty,max=100"`
	EwalletProvider      *string    `json:"ewallet_provider" validate:"omitempty,max=50"`
	EwalletNumber        *string    `json:"ewallet_number" validate:"omitempty,max=50"`
	GatewayName          *string    `json:"gateway_name" validate:"omitempty,max=50"`
	GatewayTrxID         *string    `json:"gateway_trx_id" validate:"omitempty,max=100"`
	XenditInvoiceID      *string    `json:"xendit_invoice_id" validate:"omitempty,max=100"`
	XenditExternalID     *string    `json:"xendit_external_id" validate:"omitempty,max=100"`
	XenditPaymentChannel *string    `json:"xendit_payment_channel" validate:"omitempty,max=50"`
	ProofImage           *string    `json:"proof_image"`
	ReceiptNumber        *string    `json:"receipt_number" validate:"omitempty,max=50"`
	Status               *string    `json:"status" validate:"omitempty,oneof=pending confirmed rejected refunded"`
	Notes                *string    `json:"notes"`
}

// PaymentAllocationInput for creating a payment allocation
type PaymentAllocationInput struct {
	PaymentID       uuid.UUID `json:"payment_id" validate:"required"`
	InvoiceID       uuid.UUID `json:"invoice_id" validate:"required"`
	AllocatedAmount float64   `json:"allocated_amount" validate:"required,min=0"`
}

// PaymentFilter for querying payments
type PaymentFilter struct {
	CustomerID      *uuid.UUID         `json:"customer_id"`
	InvoiceID       *uuid.UUID         `json:"invoice_id"`
	Status          *PaymentStatusType `json:"status"`
	PaymentMethod   *PaymentMethod     `json:"payment_method"`
	PaymentDateFrom *time.Time         `json:"payment_date_from"`
	PaymentDateTo   *time.Time         `json:"payment_date_to"`
	Search          *string            `json:"search"`
	IncludeDeleted  bool               `json:"include_deleted"`
}

// ToModel converts input to Payment
func (i *PaymentInput) ToModel() *Payment {
	p := &Payment{
		PaymentNumber:        i.PaymentNumber,
		CustomerID:           i.CustomerID,
		InvoiceID:            i.InvoiceID,
		Amount:               i.Amount,
		PaymentMethod:        PaymentMethod(i.PaymentMethod),
		PaymentDate:          i.PaymentDate,
		BankName:             i.BankName,
		BankAccountNumber:    i.BankAccountNumber,
		BankAccountName:      i.BankAccountName,
		TransactionReference: i.TransactionReference,
		EwalletProvider:      i.EwalletProvider,
		EwalletNumber:        i.EwalletNumber,
		GatewayName:          i.GatewayName,
		GatewayTrxID:         i.GatewayTrxID,
		XenditInvoiceID:      i.XenditInvoiceID,
		XenditExternalID:     i.XenditExternalID,
		XenditPaymentChannel: i.XenditPaymentChannel,
		ProofImage:           i.ProofImage,
		ReceiptNumber:        i.ReceiptNumber,
		Notes:                i.Notes,
	}
	if i.Status != nil {
		p.Status = PaymentStatusType(*i.Status)
	} else {
		p.Status = PaymentStatusTypePending
	}
	return p
}

func (i *PaymentAllocationInput) ToModel() *PaymentAllocation {
	return &PaymentAllocation{
		PaymentID:       i.PaymentID,
		InvoiceID:       i.InvoiceID,
		AllocatedAmount: i.AllocatedAmount,
	}
}

func (p *Payment) IsConfirmed() bool           { return p.Status == PaymentStatusTypeConfirmed }
func (p *Payment) IsPending() bool             { return p.Status == PaymentStatusTypePending }
func (p *Payment) GetRemainingAmount() float64 { return p.Amount - p.AllocatedAmount }
func (p *Payment) CanAllocate(amount float64) bool {
	return p.GetRemainingAmount() >= amount
}
