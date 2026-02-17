package model

import (
	"time"

	"github.com/google/uuid"
)

type Refund struct {
	ID               uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	RefundNumber     string     `json:"refund_number" gorm:"unique;not null"`
	PaymentID        uuid.UUID  `json:"payment_id" gorm:"type:uuid;not null"`
	Payment          *Payment   `json:"payment,omitempty" gorm:"foreignKey:PaymentID"`
	InvoiceID        *uuid.UUID `json:"invoice_id" gorm:"type:uuid"`
	Invoice          *Invoice   `json:"invoice,omitempty" gorm:"foreignKey:InvoiceID"`
	CustomerID       uuid.UUID  `json:"customer_id" gorm:"type:uuid;not null"`
	Customer         *Customer  `json:"customer,omitempty" gorm:"foreignKey:CustomerID"`
	RefundAmount     float64    `json:"refund_amount" gorm:"not null"`
	RefundType       string     `json:"refund_type" gorm:"not null"` // 'full', 'partial'
	RefundReason     *string    `json:"refund_reason" gorm:"type:text"`
	RefundMethod     *string    `json:"refund_method"`               // 'original', 'bank_transfer', 'cash'
	BankName         *string    `json:"bank_name"`
	BankAccountName  *string    `json:"bank_account_name"`
	BankAccountNumber *string   `json:"bank_account_number"`
	Status           string     `json:"status" gorm:"not null;default:'pending'"` // 'pending', 'approved', 'rejected', 'processed', 'completed', 'failed'
	ApprovedBy       *uint      `json:"approved_by" gorm:"type:integer"`
	ApprovedByUser   *User      `json:"approved_by_user,omitempty" gorm:"foreignKey:ApprovedBy"`
	ApprovedAt       *time.Time `json:"approved_at"`
	ProcessedBy      *uint      `json:"processed_by" gorm:"type:integer"`
	ProcessedByUser  *User      `json:"processed_by_user,omitempty" gorm:"foreignKey:ProcessedBy"`
	ProcessedAt      *time.Time `json:"processed_at"`
	RejectionReason  *string    `json:"rejection_reason" gorm:"type:text"`
	Notes            *string    `json:"notes" gorm:"type:text"`
	XenditRefundID   *string    `json:"xendit_refund_id"` // Xendit refund ID if using Xendit
	CreatedBy        *uint      `json:"created_by" gorm:"type:integer"`
	CreatedByUser    *User      `json:"created_by_user,omitempty" gorm:"foreignKey:CreatedBy"`
	CreatedAt        time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt        *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

func (Refund) TableName() string {
	return "refunds"
}

type RefundInput struct {
	PaymentID        uuid.UUID  `json:"payment_id" validate:"required"`
	InvoiceID        *uuid.UUID `json:"invoice_id"`
	RefundAmount     *float64   `json:"refund_amount" validate:"required,gt=0"`
	RefundType       *string    `json:"refund_type" validate:"required,oneof=full partial"`
	RefundReason     *string    `json:"refund_reason" validate:"required"`
	RefundMethod     *string    `json:"refund_method" validate:"required,oneof=original bank_transfer cash"`
	BankName         *string    `json:"bank_name"`
	BankAccountName  *string    `json:"bank_account_name"`
	BankAccountNumber *string   `json:"bank_account_number"`
	Notes            *string    `json:"notes"`
}

type RefundFilter struct {
	IDs            []uuid.UUID `json:"ids"`
	RefundNumbers  []string    `json:"refund_numbers"`
	PaymentIDs     []uuid.UUID `json:"payment_ids"`
	InvoiceIDs     []uuid.UUID `json:"invoice_ids"`
	CustomerIDs    []uuid.UUID `json:"customer_ids"`
	CustomerID     *uuid.UUID  `json:"customer_id"`
	Status         []string    `json:"status"`
	RefundType     []string    `json:"refund_type"`
	RefundMethod   []string    `json:"refund_method"`
	CreatedStart   *time.Time  `json:"created_start"`
	CreatedEnd     *time.Time  `json:"created_end"`
	ApprovedStart  *time.Time  `json:"approved_start"`
	ApprovedEnd    *time.Time  `json:"approved_end"`
	ProcessedStart *time.Time  `json:"processed_start"`
	ProcessedEnd   *time.Time  `json:"processed_end"`
	AmountMin      *float64    `json:"amount_min"`
	AmountMax      *float64    `json:"amount_max"`
	Search         *string     `json:"search"`
}

func (f RefundFilter) IsEmpty() bool {
	return len(f.IDs) == 0 &&
		len(f.RefundNumbers) == 0 &&
		len(f.PaymentIDs) == 0 &&
		len(f.InvoiceIDs) == 0 &&
		len(f.CustomerIDs) == 0 &&
		f.CustomerID == nil &&
		len(f.Status) == 0 &&
		len(f.RefundType) == 0 &&
		len(f.RefundMethod) == 0 &&
		f.CreatedStart == nil &&
		f.CreatedEnd == nil &&
		f.ApprovedStart == nil &&
		f.ApprovedEnd == nil &&
		f.ProcessedStart == nil &&
		f.ProcessedEnd == nil &&
		f.AmountMin == nil &&
		f.AmountMax == nil &&
		f.Search == nil
}

func RefundPrepare(refund *Refund) {
	if refund.RefundNumber == "" {
		refund.RefundNumber = GenerateRefundNumber()
	}
	if refund.Status == "" {
		refund.Status = "pending"
	}
}

func GenerateRefundNumber() string {
	// Format: RFD-YYYYMMDD-XXXX
	timestamp := time.Now().Format("20060102")
	randomPart := uuid.New().String()[:4]
	return "RFD-" + timestamp + "-" + randomPart
}
