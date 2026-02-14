package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// PaymentMethod represents the payment method
type PaymentMethod string

const (
	PaymentMethodVA           PaymentMethod = "va"
	PaymentMethodEWallet      PaymentMethod = "e-wallet"
	PaymentMethodCreditCard   PaymentMethod = "credit_card"
	PaymentMethodDebitCard    PaymentMethod = "debit_card"
	PaymentMethodQRIS         PaymentMethod = "qris"
	PaymentMethodCash         PaymentMethod = "cash"
	PaymentMethodBankTransfer PaymentMethod = "bank_transfer"
)

// PaymentStatusType represents the status of a payment
type PaymentStatusType string

const (
	PaymentStatusPending   PaymentStatusType = "pending"
	PaymentStatusConfirmed PaymentStatusType = "confirmed"
	PaymentStatusRejected  PaymentStatusType = "rejected"
	PaymentStatusRefunded  PaymentStatusType = "refunded"
)

type Payment struct {
	ID                   uuid.UUID         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PaymentNumber        string            `json:"payment_number" gorm:"uniqueIndex;not null"`
	CustomerID           uuid.UUID         `json:"customer_id" gorm:"type:uuid;not null;index"`
	Customer             *Customer         `json:"customer,omitempty" gorm:"foreignKey:CustomerID;references:ID"`
	InvoiceID            *uuid.UUID        `json:"invoice_id" gorm:"type:uuid;index"`
	Invoice              *Invoice          `json:"invoice,omitempty" gorm:"foreignKey:InvoiceID;references:ID"`
	Amount               decimal.Decimal   `json:"amount" gorm:"type:decimal(12,2);not null;default:0"`
	AllocatedAmount      decimal.Decimal   `json:"allocated_amount" gorm:"type:decimal(12,2);not null;default:0"`
	PaymentMethod        PaymentMethod     `json:"payment_method" gorm:"type:payment_method;not null;index"`
	PaymentDate          time.Time         `json:"payment_date" gorm:"not null;index"`
	BankName             string            `json:"bank_name"`
	BankAccountNumber    string            `json:"bank_account_number"`
	BankAccountName      string            `json:"bank_account_name"`
	TransactionReference string            `json:"transaction_reference" gorm:"index"`
	EWalletProvider      string            `json:"ewallet_provider"`
	EWalletNumber        string            `json:"ewallet_number"`
	ProofImage           string            `json:"proof_image"`
	ReceiptNumber        string            `json:"receipt_number"`
	Status               PaymentStatusType `json:"status" gorm:"type:payment_status_type;not null;default:'pending';index"`
	ProcessedBy          *uint             `json:"processed_by" gorm:"index"`
	ProcessedByUser      *User             `json:"processed_by_user,omitempty" gorm:"foreignKey:ProcessedBy;references:ID"`
	ProcessedAt          *time.Time        `json:"processed_at"`
	RejectionReason      string            `json:"rejection_reason"`
	RefundAmount         decimal.Decimal   `json:"refund_amount" gorm:"type:decimal(12,2);default:0"`
	RefundDate           *time.Time        `json:"refund_date"`
	RefundReason         string            `json:"refund_reason"`
	RefundedBy           *uint             `json:"refunded_by"`
	RefundedByUser       *User             `json:"refunded_by_user,omitempty" gorm:"foreignKey:RefundedBy;references:ID"`
	Notes                string            `json:"notes"`
	Allocations          []PaymentAllocation `json:"allocations,omitempty" gorm:"foreignKey:PaymentID;constraint:OnDelete:CASCADE"`
	CreatedBy            *uint             `json:"created_by" gorm:"index"`
	CreatedByUser        *User             `json:"created_by_user,omitempty" gorm:"foreignKey:CreatedBy;references:ID"`
	CreatedAt            time.Time         `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt            time.Time         `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt            *time.Time        `json:"deleted_at,omitempty" gorm:"index"`
}

type PaymentAllocation struct {
	ID              uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PaymentID       uuid.UUID       `json:"payment_id" gorm:"type:uuid;not null;index"`
	InvoiceID       uuid.UUID       `json:"invoice_id" gorm:"type:uuid;not null;index"`
	Invoice         *Invoice        `json:"invoice,omitempty" gorm:"foreignKey:InvoiceID;references:ID"`
	AllocatedAmount decimal.Decimal `json:"allocated_amount" gorm:"type:decimal(12,2);not null;default:0"`
	CreatedAt       time.Time       `json:"created_at" gorm:"autoCreateTime"`
}

type PaymentInput struct {
	CustomerID           uuid.UUID     `json:"customer_id" binding:"required"`
	InvoiceID            *uuid.UUID    `json:"invoice_id"`
	Amount               string        `json:"amount" binding:"required"`
	PaymentMethod        PaymentMethod `json:"payment_method" binding:"required,oneof=va e-wallet credit_card debit_card qris cash bank_transfer"`
	PaymentDate          time.Time     `json:"payment_date" binding:"required"`
	BankName             string        `json:"bank_name"`
	BankAccountNumber    string        `json:"bank_account_number"`
	BankAccountName      string        `json:"bank_account_name"`
	TransactionReference string        `json:"transaction_reference"`
	EWalletProvider      string        `json:"ewallet_provider"`
	EWalletNumber        string        `json:"ewallet_number"`
	ProofImage           string        `json:"proof_image"`
	ReceiptNumber        string        `json:"receipt_number"`
	Notes                string        `json:"notes"`
}

type PaymentFilter struct {
	IDs                   []uuid.UUID         `json:"ids"`
	PaymentNumbers        []string            `json:"payment_numbers"`
	CustomerIDs           []uuid.UUID         `json:"customer_ids"`
	InvoiceIDs            []uuid.UUID         `json:"invoice_ids"`
	Statuses              []PaymentStatusType `json:"statuses"`
	PaymentMethods        []PaymentMethod     `json:"payment_methods"`
	TransactionReferences []string            `json:"transaction_references"`
	PaymentDateBefore     *time.Time          `json:"payment_date_before"`
	PaymentDateAfter      *time.Time          `json:"payment_date_after"`
	Search                string              `json:"search"`
	Page                  int                 `json:"page"`
	Limit                 int                 `json:"limit"`
}

func (f PaymentFilter) IsEmpty() bool {
	return len(f.IDs) == 0 && len(f.PaymentNumbers) == 0 && len(f.CustomerIDs) == 0 &&
		len(f.InvoiceIDs) == 0 && len(f.Statuses) == 0 && len(f.PaymentMethods) == 0 &&
		len(f.TransactionReferences) == 0 && f.PaymentDateBefore == nil &&
		f.PaymentDateAfter == nil && f.Search == ""
}

// TableName specifies the table name for GORM
func (Payment) TableName() string {
	return "payments"
}

// TableName specifies the table name for GORM
func (PaymentAllocation) TableName() string {
	return "payment_allocations"
}

// UnallocatedAmount returns the amount that hasn't been allocated yet
func (p *Payment) UnallocatedAmount() decimal.Decimal {
	return p.Amount.Sub(p.AllocatedAmount)
}

// IsFullyAllocated checks if all payment amount has been allocated
func (p *Payment) IsFullyAllocated() bool {
	return p.AllocatedAmount.GreaterThanOrEqual(p.Amount)
}

// GeneratePaymentNumber generates a unique payment number
func GeneratePaymentNumber(year int, month int, sequence int) string {
	prefix := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC).Format("PAY/2006/01/")
	seqStr := ""
	num := sequence

	// Convert number to string
	if num == 0 {
		seqStr = "0"
	} else {
		for num > 0 {
			digit := num % 10
			seqStr = string(rune('0'+digit)) + seqStr
			num /= 10
		}
	}

	// Pad left with zeros
	for len(seqStr) < 4 {
		seqStr = "0" + seqStr
	}

	return prefix + seqStr
}

// XenditWebhookData represents webhook data from Xendit
type XenditWebhookData struct {
	ID                 string    `json:"id"`
	ExternalID         string    `json:"external_id"`
	UserID             string    `json:"user_id"`
	Status             string    `json:"status"`
	Amount             float64   `json:"amount"`
	PaidAmount         float64   `json:"paid_amount"`
	BankCode           string    `json:"bank_code"`
	AccountNumber      string    `json:"account_number"`
	MerchantName       string    `json:"merchant_name"`
	PaymentMethod      string    `json:"payment_method"`
	PaymentChannel     string    `json:"payment_channel"`
	PaymentDestination string    `json:"payment_destination"`
	Description        string    `json:"description"`
	TransactionID      string    `json:"transaction_id"`
	PaymentID          string    `json:"payment_id"`
	PaidAt             time.Time `json:"paid_at"`
	Created            time.Time `json:"created"`
	Updated            time.Time `json:"updated"`
	Currency           string    `json:"currency"`
	CallbackVirtualAccountID string `json:"callback_virtual_account_id"`
}
