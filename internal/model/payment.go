package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PaymentMethod represents payment method type
type PaymentMethod string

const (
	PaymentMethodCash        PaymentMethod = "cash"
	PaymentMethodBankTransfer PaymentMethod = "bank_transfer"
	PaymentMethodEWallet     PaymentMethod = "e-wallet"
	PaymentMethodCreditCard  PaymentMethod = "credit_card"
	PaymentMethodDebitCard   PaymentMethod = "debit_card"
	PaymentMethodCheck       PaymentMethod = "check"
	PaymentMethodXendit      PaymentMethod = "xendit"
)

// PaymentStatusType represents payment status
type PaymentStatusType string

const (
	PaymentStatusTypePending   PaymentStatusType = "pending"
	PaymentStatusTypeConfirmed PaymentStatusType = "confirmed"
	PaymentStatusTypeRejected  PaymentStatusType = "rejected"
	PaymentStatusTypeRefunded  PaymentStatusType = "refunded"
)

// Payment represents customer payment
type Payment struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PaymentNumber string         `gorm:"size:50;uniqueIndex;not null" json:"payment_number" validate:"required"`

	CustomerID uuid.UUID `gorm:"type:uuid;not null" json:"customer_id" validate:"required"`
	Customer   *Customer `gorm:"foreignKey:CustomerID;constraint:OnDelete:RESTRICT" json:"customer,omitempty"`

	InvoiceID *uuid.UUID `gorm:"type:uuid" json:"invoice_id"` // nullable for advance payment
	Invoice   *Invoice   `gorm:"foreignKey:InvoiceID;constraint:OnDelete:RESTRICT" json:"invoice,omitempty"`

	// Amount
	Amount           float64 `gorm:"type:decimal(12,2);not null" json:"amount" validate:"required,min=0"`
	AllocatedAmount  float64 `gorm:"type:decimal(12,2);default:0" json:"allocated_amount"`
	RemainingAmount  float64 `gorm:"type:decimal(12,2);->;default:0" json:"remaining_amount"` // computed column

	// Payment Details
	PaymentMethod PaymentMethod `gorm:"size:20;not null" json:"payment_method" validate:"required"`
	PaymentDate   time.Time     `gorm:"not null" json:"payment_date" validate:"required"`

	// Bank Transfer Details
	BankName           *string `gorm:"size:100" json:"bank_name"`
	BankAccountNumber  *string `gorm:"size:50" json:"bank_account_number"`
	BankAccountName    *string `gorm:"size:100" json:"bank_account_name"`
	TransactionReference *string `gorm:"size:100" json:"transaction_reference"`

	// E-wallet Details
	EwalletProvider *string `gorm:"size:50" json:"ewallet_provider"` // gopay, ovo, dana, dll
	EwalletNumber   *string `gorm:"size:50" json:"ewallet_number"`

	// Xendit Details
	XenditInvoiceID    *string `gorm:"size:100" json:"xendit_invoice_id"`
	XenditExternalID   *string `gorm:"size:100" json:"xendit_external_id"`
	XenditPaymentChannel *string `gorm:"size:50" json:"xendit_payment_channel"`

	// Proof
	ProofImage    *string `gorm:"type:text" json:"proof_image"` // URL/path
	ReceiptNumber *string `gorm:"size:50" json:"receipt_number"`

	// Status
	Status PaymentStatusType `gorm:"size:20;not null;default:pending" json:"status"`

	// Processing
	ProcessedBy     *uint      `gorm:"type:bigint" json:"processed_by"`
	ProcessedAt     *time.Time `json:"processed_at"`
	RejectionReason *string    `gorm:"type:text" json:"rejection_reason"`

	// Refund
	RefundAmount *float64   `gorm:"type:decimal(12,2);default:0" json:"refund_amount"`
	RefundDate   *time.Time `json:"refund_date"`
	RefundReason *string    `gorm:"type:text" json:"refund_reason"`
	RefundedBy   *uint      `gorm:"type:bigint" json:"refunded_by"`

	Notes *string `gorm:"type:text" json:"notes"`

	// Relations
	Allocations []PaymentAllocation `gorm:"foreignKey:PaymentID" json:"allocations,omitempty"`

	CreatedBy *uuid.UUID `gorm:"type:uuid" json:"created_by"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for Payment
func (Payment) TableName() string {
	return "payments"
}

// PaymentAllocation represents allocation of payment to invoice
type PaymentAllocation struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PaymentID       uuid.UUID `gorm:"type:uuid;not null" json:"payment_id" validate:"required"`
	Payment         *Payment  `gorm:"foreignKey:PaymentID;constraint:OnDelete:CASCADE" json:"payment,omitempty"`
	InvoiceID       uuid.UUID `gorm:"type:uuid;not null" json:"invoice_id" validate:"required"`
	Invoice         *Invoice  `gorm:"foreignKey:InvoiceID;constraint:OnDelete:CASCADE" json:"invoice,omitempty"`
	AllocatedAmount float64   `gorm:"type:decimal(12,2);not null" json:"allocated_amount" validate:"required,min=0"`

	CreatedAt time.Time `json:"created_at"`
}

// TableName specifies the table name for PaymentAllocation
func (PaymentAllocation) TableName() string {
	return "payment_allocations"
}

// PaymentInput for creating/updating payment
type PaymentInput struct {
	PaymentNumber        string     `json:"payment_number" validate:"required,max=50"`
	CustomerID           uuid.UUID  `json:"customer_id" validate:"required,uuid"`
	InvoiceID            *uuid.UUID `json:"invoice_id" validate:"omitempty,uuid"`
	Amount               float64    `json:"amount" validate:"required,min=0"`
	PaymentMethod        string     `json:"payment_method" validate:"required,oneof=cash bank_transfer e-wallet credit_card debit_card check xendit"`
	PaymentDate          time.Time  `json:"payment_date" validate:"required"`
	BankName             *string    `json:"bank_name" validate:"omitempty,max=100"`
	BankAccountNumber    *string    `json:"bank_account_number" validate:"omitempty,max=50"`
	BankAccountName      *string    `json:"bank_account_name" validate:"omitempty,max=100"`
	TransactionReference *string    `json:"transaction_reference" validate:"omitempty,max=100"`
	EwalletProvider      *string    `json:"ewallet_provider" validate:"omitempty,max=50"`
	EwalletNumber        *string    `json:"ewallet_number" validate:"omitempty,max=50"`
	XenditInvoiceID      *string    `json:"xendit_invoice_id" validate:"omitempty,max=100"`
	XenditExternalID     *string    `json:"xendit_external_id" validate:"omitempty,max=100"`
	XenditPaymentChannel *string    `json:"xendit_payment_channel" validate:"omitempty,max=50"`
	ProofImage           *string    `json:"proof_image"`
	ReceiptNumber        *string    `json:"receipt_number" validate:"omitempty,max=50"`
	Status               *string    `json:"status" validate:"omitempty,oneof=pending confirmed rejected refunded"`
	Notes                *string    `json:"notes"`
}

// PaymentAllocationInput for creating payment allocation
type PaymentAllocationInput struct {
	PaymentID       uuid.UUID `json:"payment_id" validate:"required,uuid"`
	InvoiceID       uuid.UUID `json:"invoice_id" validate:"required,uuid"`
	AllocatedAmount float64   `json:"allocated_amount" validate:"required,min=0"`
}

// PaymentFilter for filtering payments
type PaymentFilter struct {
	CustomerID     *uuid.UUID         `json:"customer_id"`
	InvoiceID      *uuid.UUID         `json:"invoice_id"`
	Status         *PaymentStatusType `json:"status"`
	PaymentMethod  *PaymentMethod     `json:"payment_method"`
	PaymentDateFrom *time.Time        `json:"payment_date_from"`
	PaymentDateTo   *time.Time        `json:"payment_date_to"`
	Search         *string            `json:"search"` // search by payment number, reference
	IncludeDeleted bool               `json:"include_deleted"`
}

// ToModel converts input to model
func (i *PaymentInput) ToModel() *Payment {
	payment := &Payment{
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
		XenditInvoiceID:      i.XenditInvoiceID,
		XenditExternalID:     i.XenditExternalID,
		XenditPaymentChannel: i.XenditPaymentChannel,
		ProofImage:           i.ProofImage,
		ReceiptNumber:        i.ReceiptNumber,
		Notes:                i.Notes,
	}

	// Set status
	if i.Status != nil {
		payment.Status = PaymentStatusType(*i.Status)
	} else {
		payment.Status = PaymentStatusTypePending
	}

	return payment
}

// ToModel converts input to model
func (i *PaymentAllocationInput) ToModel() *PaymentAllocation {
	return &PaymentAllocation{
		PaymentID:       i.PaymentID,
		InvoiceID:       i.InvoiceID,
		AllocatedAmount: i.AllocatedAmount,
	}
}

// IsConfirmed checks if payment is confirmed
func (p *Payment) IsConfirmed() bool {
	return p.Status == PaymentStatusTypeConfirmed
}

// IsPending checks if payment is pending
func (p *Payment) IsPending() bool {
	return p.Status == PaymentStatusTypePending
}

// GetRemainingAmount returns remaining unallocated amount
func (p *Payment) GetRemainingAmount() float64 {
	return p.Amount - p.AllocatedAmount
}

// CanAllocate checks if payment can allocate more
func (p *Payment) CanAllocate(amount float64) bool {
	return p.GetRemainingAmount() >= amount
}
