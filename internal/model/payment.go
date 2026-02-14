package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Payment struct {
	ID                   uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PaymentNumber        string     `json:"payment_number" gorm:"unique;not null;size:50" validate:"required"`
	CustomerID           uuid.UUID  `json:"customer_id" gorm:"type:uuid;not null" validate:"required"`
	Customer             *Customer  `json:"customer,omitempty" gorm:"foreignKey:CustomerID"`
	InvoiceID            *uuid.UUID `json:"invoice_id" gorm:"type:uuid" validate:"omitempty"`
	Invoice              *Invoice   `json:"invoice,omitempty" gorm:"foreignKey:InvoiceID"`
	Amount               float64    `json:"amount" gorm:"not null;type:decimal(12,2)" validate:"required,min=0"`
	AllocatedAmount      float64    `json:"allocated_amount" gorm:"type:decimal(12,2);default:0" validate:"min=0"`
	PaymentMethod        string     `json:"payment_method" gorm:"not null;type:varchar(20)" validate:"required,oneof=va e-wallet credit_card debit_card check"`
	PaymentDate          time.Time  `json:"payment_date" gorm:"not null;type:timestamp" validate:"required"`
	BankName             *string    `json:"bank_name" gorm:"size:100"`
	BankAccountNumber    *string    `json:"bank_account_number" gorm:"size:50"`
	BankAccountName      *string    `json:"bank_account_name" gorm:"size:100"`
	TransactionReference *string    `json:"transaction_reference" gorm:"size:100"`
	EwalletProvider      *string    `json:"ewallet_provider" gorm:"size:20" validate:"omitempty,oneof=gopay ovo dana"`
	EwalletNumber        *string    `json:"ewallet_number" gorm:"size:30"`
	ProofImage           *string    `json:"proof_image" gorm:"type:text"`
	ReceiptNumber        *string    `json:"receipt_number" gorm:"size:50"`
	Status               string     `json:"status" gorm:"not null;default:'pending';type:varchar(20)" validate:"required,oneof=pending confirmed rejected refunded"`
	ProcessedBy          *uuid.UUID `json:"processed_by" gorm:"type:uuid"`
	ProcessedByUser      *User      `json:"processed_by_user,omitempty" gorm:"foreignKey:ProcessedBy"`
	ProcessedAt          *time.Time `json:"processed_at" gorm:"type:timestamp"`
	RejectionReason      *string    `json:"rejection_reason" gorm:"type:text"`
	RefundAmount         float64    `json:"refund_amount" gorm:"type:decimal(12,2);default:0" validate:"min=0"`
	RefundDate           *time.Time `json:"refund_date" gorm:"type:timestamp"`
	RefundReason         *string    `json:"refund_reason" gorm:"type:text"`
	RefundedBy           *uuid.UUID `json:"refunded_by" gorm:"type:uuid"`
	Notes                *string    `json:"notes" gorm:"type:text"`
	CreatedBy            *uuid.UUID `json:"created_by" gorm:"type:uuid"`
	CreatedByUser        *User      `json:"created_by_user,omitempty" gorm:"foreignKey:CreatedBy"`
	CreatedAt            time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt            time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt            *time.Time `json:"-" gorm:"index"`
}

type PaymentInput struct {
	PaymentNumber        *string    `json:"payment_number" validate:"omitempty,required_if=Create"`
	CustomerID           uuid.UUID  `json:"customer_id" validate:"required"`
	InvoiceID            *uuid.UUID `json:"invoice_id"`
	Amount               *float64   `json:"amount" validate:"required,min=0"`
	PaymentMethod        *string    `json:"payment_method" validate:"required,oneof=va e-wallet credit_card debit_card check"`
	PaymentDate          *time.Time `json:"payment_date" validate:"required"`
	BankName             *string    `json:"bank_name" validate:"omitempty,max=100"`
	BankAccountNumber    *string    `json:"bank_account_number" validate:"omitempty,max=50"`
	BankAccountName      *string    `json:"bank_account_name" validate:"omitempty,max=100"`
	TransactionReference *string    `json:"transaction_reference" validate:"omitempty,max=100"`
	EwalletProvider      *string    `json:"ewallet_provider" validate:"omitempty,oneof=gopay ovo dana"`
	EwalletNumber        *string    `json:"ewallet_number" validate:"omitempty,max=30"`
	ProofImage           *string    `json:"proof_image"`
	ReceiptNumber        *string    `json:"receipt_number" validate:"omitempty,max=50"`
	Notes                *string    `json:"notes"`
}

type PaymentFilter struct {
	IDs             []uuid.UUID `json:"ids"`
	PaymentNumbers  []string    `json:"payment_numbers"`
	CustomerIDs     []uuid.UUID `json:"customer_ids"`
	InvoiceIDs      []uuid.UUID `json:"invoice_ids"`
	Status          []string    `json:"status"`
	PaymentMethod   []string    `json:"payment_method"`
	PaymentStart    *time.Time  `json:"payment_start"`
	PaymentEnd      *time.Time  `json:"payment_end"`
	AmountMin       *float64    `json:"amount_min"`
	AmountMax       *float64    `json:"amount_max"`
	EwalletProvider []string    `json:"ewallet_provider"`
	TransactionRef  *string     `json:"transaction_ref"`
	Search          *string     `json:"search"`
	IsProcessed     *bool       `json:"is_processed"`
	IsRefunded      *bool       `json:"is_refunded"`
}

type PaymentAllocation struct {
	ID              uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PaymentID       uuid.UUID `json:"payment_id" gorm:"type:uuid;not null;index:idx_payment_allocations_payment,priority:1" validate:"required"`
	InvoiceID       uuid.UUID `json:"invoice_id" gorm:"type:uuid;not null;index:idx_payment_allocations_invoice,priority:1" validate:"required"`
	Payment         *Payment  `json:"payment,omitempty" gorm:"foreignKey:PaymentID"`
	Invoice         *Invoice  `json:"invoice,omitempty" gorm:"foreignKey:InvoiceID"`
	AllocatedAmount float64   `json:"allocated_amount" gorm:"not null;type:decimal(12,2)" validate:"required,min=0"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
}

type PaymentAllocationInput struct {
	PaymentID       uuid.UUID `json:"payment_id" validate:"required"`
	InvoiceID       uuid.UUID `json:"invoice_id" validate:"required"`
	AllocatedAmount *float64  `json:"allocated_amount" validate:"required,min=0"`
}

func (f PaymentFilter) IsEmpty() bool {
	return len(f.IDs) == 0 && len(f.PaymentNumbers) == 0 &&
		len(f.CustomerIDs) == 0 && len(f.InvoiceIDs) == 0 &&
		len(f.Status) == 0 && len(f.PaymentMethod) == 0 &&
		f.PaymentStart == nil && f.PaymentEnd == nil &&
		f.AmountMin == nil && f.AmountMax == nil &&
		len(f.EwalletProvider) == 0 && f.TransactionRef == nil &&
		f.Search == nil && f.IsProcessed == nil && f.IsRefunded == nil
}

func PaymentPrepare(v *Payment) {
	if v.Status == "" {
		v.Status = "pending"
	}
	if v.PaymentNumber == "" {
		v.PaymentNumber = generatePaymentNumber()
	}
}

func generatePaymentNumber() string {
	return fmt.Sprintf("PAY-%s-%04d", time.Now().Format("20060102"), time.Now().UnixNano()%10000)
}
