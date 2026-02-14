package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type CashCategory struct {
	ID               uuid.UUID     `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Code             string        `json:"code" gorm:"unique;not null;size:50" validate:"required"`
	Name             string        `json:"name" gorm:"not null;size:200" validate:"required"`
	Type             string        `json:"type" gorm:"not null;type:varchar(20)" validate:"required,oneof=income expense"`
	ParentCategoryID *uuid.UUID    `json:"parent_category_id" gorm:"type:uuid"`
	ParentCategory   *CashCategory `json:"parent_category,omitempty" gorm:"foreignKey:ParentCategoryID"`
	Description      *string       `json:"description" gorm:"type:text"`
	IsSystem         bool          `json:"is_system" gorm:"not null;default:false" validate:"required"`
	IsActive         bool          `json:"is_active" gorm:"not null;default:true" validate:"required"`
	SortOrder        int           `json:"sort_order" gorm:"default:0" validate:"required"`
	CreatedAt        time.Time     `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time     `json:"updated_at" gorm:"autoUpdateTime"`
}

type CashCategoryInput struct {
	Code             *string    `json:"code" validate:"omitempty,required_if=Create"`
	Name             *string    `json:"name" validate:"omitempty,required_if=Create"`
	Type             *string    `json:"type" validate:"omitempty,required,oneof=income expense"`
	ParentCategoryID *uuid.UUID `json:"parent_category_id"`
	Description      *string    `json:"description"`
	IsSystem         *bool      `json:"is_system"`
	IsActive         *bool      `json:"is_active"`
	SortOrder        *int       `json:"sort_order"`
}

type CashCategoryFilter struct {
	IDs      []uuid.UUID `json:"ids"`
	Codes    []string    `json:"codes"`
	Type     *string     `json:"type" validate:"omitempty,oneof=income expense"`
	IsActive *bool       `json:"is_active"`
	IsSystem *bool       `json:"is_system"`
	Search   *string     `json:"search"`
}

func (f CashCategoryFilter) IsEmpty() bool {
	return len(f.IDs) == 0 && len(f.Codes) == 0 &&
		f.Type == nil && f.IsActive == nil && f.IsSystem == nil && f.Search == nil
}

type CashTransaction struct {
	ID                uuid.UUID     `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TransactionNumber string        `json:"transaction_number" gorm:"unique;not null;size:50" validate:"required"`
	TransactionDate   time.Time     `json:"transaction_date" gorm:"not null;type:timestamp" validate:"required"`
	Type              string        `json:"type" gorm:"not null;type:varchar(20)" validate:"required,oneof=income expense"`
	CategoryID        uuid.UUID     `json:"category_id" gorm:"type:uuid;not null" validate:"required"`
	Category          *CashCategory `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	Amount            float64       `json:"amount" gorm:"not null;type:decimal(12,2)" validate:"required,min=0"`
	PaymentMethod     *string       `json:"payment_method" gorm:"size:50"`
	Description       *string       `json:"description" gorm:"type:text"`
	ReferenceType     *string       `json:"reference_type" gorm:"type:varchar(50)" validate:"omitempty,oneof=payment invoice expense other none"`
	ReferenceID       *uuid.UUID    `json:"reference_id" gorm:"type:uuid"`
	CustomerID        *uuid.UUID    `json:"customer_id" gorm:"type:uuid"`
	Customer          *Customer     `json:"customer,omitempty" gorm:"foreignKey:CustomerID"`
	AccountName       *string       `json:"account_name" gorm:"size:100"`
	AccountNumber     *string       `json:"account_number" gorm:"size:50"`
	ProofImage        *string       `json:"proof_image" gorm:"type:text"`
	ReceiptNumber     *string       `json:"receipt_number" gorm:"size:50"`
	RequiresApproval  bool          `json:"requires_approval" gorm:"default:false" validate:"required"`
	ApprovalStatus    string        `json:"approval_status" gorm:"default:'pending';type:varchar(20)" validate:"required,oneof=pending approved rejected"`
	ApprovedBy        *uuid.UUID    `json:"approved_by" gorm:"type:uuid"`
	ApprovedByUser    *User         `json:"approved_by_user,omitempty" gorm:"foreignKey:ApprovedBy"`
	ApprovedAt        *time.Time    `json:"approved_at" gorm:"type:timestamp"`
	Notes             *string       `json:"notes" gorm:"type:text"`
	ProcessedBy       *uuid.UUID    `json:"processed_by" gorm:"type:uuid"`
	CreatedAt         time.Time     `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time     `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt         *time.Time    `json:"-" gorm:"index"`
}

type CashTransactionInput struct {
	TransactionNumber *string    `json:"transaction_number" validate:"omitempty,required_if=Create"`
	TransactionDate   *time.Time `json:"transaction_date" validate:"omitempty,required"`
	Type              *string    `json:"type" validate:"omitempty,required,oneof=income expense"`
	CategoryID        uuid.UUID  `json:"category_id" validate:"required"`
	Amount            *float64   `json:"amount" validate:"omitempty,required,min=0"`
	PaymentMethod     *string    `json:"payment_method"`
	Description       *string    `json:"description"`
	ReferenceType     *string    `json:"reference_type" validate:"omitempty,oneof=payment invoice expense other none"`
	ReferenceID       *uuid.UUID `json:"reference_id"`
	CustomerID        *uuid.UUID `json:"customer_id"`
	AccountName       *string    `json:"account_name"`
	AccountNumber     *string    `json:"account_number"`
	ProofImage        *string    `json:"proof_image"`
	ReceiptNumber     *string    `json:"receipt_number"`
	RequiresApproval  *bool      `json:"requires_approval"`
	Notes             *string    `json:"notes"`
}

type CashTransactionFilter struct {
	IDs               []uuid.UUID `json:"ids"`
	TransactionNumber []string    `json:"transaction_numbers"`
	Type              *string     `json:"type" validate:"omitempty,oneof=income expense"`
	CategoryIDs       []uuid.UUID `json:"category_ids"`
	ReferenceType     *string     `json:"reference_type" validate:"omitempty,oneof=payment invoice expense other none"`
	ReferenceID       *uuid.UUID  `json:"reference_id"`
	CustomerIDs       []uuid.UUID `json:"customer_ids"`
	DateStart         *time.Time  `json:"date_start"`
	DateEnd           *time.Time  `json:"date_end"`
	AmountMin         *float64    `json:"amount_min" validate:"omitempty,min=0"`
	AmountMax         *float64    `json:"amount_max" validate:"omitempty,min=0"`
	ApprovalStatus    *string     `json:"approval_status" validate:"omitempty,oneof=pending approved rejected"`
	RequiresApproval  *bool       `json:"requires_approval"`
	Search            *string     `json:"search"`
}

func (f CashTransactionFilter) IsEmpty() bool {
	return len(f.IDs) == 0 && len(f.TransactionNumber) == 0 &&
		f.Type == nil && len(f.CategoryIDs) == 0 &&
		f.ReferenceType == nil && f.ReferenceID == nil &&
		len(f.CustomerIDs) == 0 && f.DateStart == nil && f.DateEnd == nil &&
		f.AmountMin == nil && f.AmountMax == nil &&
		f.ApprovalStatus == nil && f.RequiresApproval == nil && f.Search == nil
}

func CashTransactionPrepare(v *CashTransaction) {
	if v.TransactionNumber == "" {
		v.TransactionNumber = generateCashTransactionNumber()
	}
	if v.ApprovalStatus == "" {
		v.ApprovalStatus = "pending"
	}
}

func generateCashTransactionNumber() string {
	return fmt.Sprintf("TRX-%s-%04d", time.Now().Format("20060102"), time.Now().UnixNano()%10000)
}
