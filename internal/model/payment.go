package model

import "time"

const (
	PaymentStatusPending  = "pending"
	PaymentStatusVerified = "verified"
	PaymentStatusRejected = "rejected"
)

type Payment struct {
	ID string `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	PaymentInput
	Invoice       *Invoice       `json:"invoice,omitempty" gorm:"foreignKey:InvoiceID"`
	PaymentMethod *PaymentMethod `json:"payment_method,omitempty" gorm:"foreignKey:PaymentMethodID"`
}

func (Payment) TableName() string {
	return "payments"
}

type PaymentInput struct {
	TenantID        string     `json:"tenant_id" gorm:"column:tenant_id;type:uuid;not null;index"`
	InvoiceID       string     `json:"invoice_id" gorm:"column:invoice_id;type:uuid;not null"`
	PaymentMethodID string     `json:"payment_method_id" gorm:"column:payment_method_id;type:uuid;not null"`
	Amount          int64      `json:"amount" gorm:"column:amount;not null"`
	PaymentDate     time.Time  `json:"payment_date" gorm:"column:payment_date;not null"`
	ProofURL        string     `json:"proof_url" gorm:"column:proof_url;type:varchar(500)"`
	Status          string     `json:"status" gorm:"column:status;type:varchar(50);default:'pending'"`
	VerifiedBy      *string    `json:"verified_by" gorm:"column:verified_by;type:uuid"`
	VerifiedAt      *time.Time `json:"verified_at" gorm:"column:verified_at"`
	Notes           string     `json:"notes" gorm:"column:notes;type:text"`
	CreatedAt       time.Time  `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       time.Time  `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

type PaymentFilter struct {
	IDs               []string `json:"ids"`
	TenantIDs         []string `json:"tenant_ids"`
	InvoiceIDs        []string `json:"invoice_ids"`
	PaymentMethodIDs  []string `json:"payment_method_ids"`
	Statuses          []string `json:"statuses"`
	VerifiedByIDs     []string `json:"verified_by_ids"`
	WithTenant        bool     `json:"-"`
	WithInvoice       bool     `json:"-"`
	WithPaymentMethod bool     `json:"-"`
	WithVerifiedBy    bool     `json:"-"`
}

func (f PaymentFilter) IsEmpty() bool {
	return len(f.IDs) == 0 && len(f.TenantIDs) == 0 && len(f.InvoiceIDs) == 0 &&
		len(f.PaymentMethodIDs) == 0 && len(f.Statuses) == 0
}
