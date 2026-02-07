package model

import "time"

const (
	InvoiceStatusUnpaid    = "unpaid"
	InvoiceStatusPaid      = "paid"
	InvoiceStatusOverdue   = "overdue"
	InvoiceStatusCancelled = "cancelled"
	InvoiceStatusPartial   = "partial"
)

type Invoice struct {
	ID string `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	InvoiceInput
	Customer     *Customer     `json:"customer,omitempty" gorm:"foreignKey:CustomerID"`
	Subscription *Subscription `json:"subscription,omitempty" gorm:"foreignKey:SubscriptionID"`
}

func (Invoice) TableName() string {
	return "invoices"
}

type InvoiceInput struct {
	TenantID       string     `json:"tenant_id" gorm:"column:tenant_id;type:uuid;not null;index"`
	CustomerID     string     `json:"customer_id" gorm:"column:customer_id;type:uuid;not null"`
	SubscriptionID string     `json:"subscription_id" gorm:"column:subscription_id;type:uuid;not null"`
	InvoiceNumber  string     `json:"invoice_number" gorm:"column:invoice_number;type:varchar(100);uniqueIndex;not null"`
	Amount         int64      `json:"amount" gorm:"column:amount;not null"`
	TaxAmount      int64      `json:"tax_amount" gorm:"column:tax_amount;default:0"`
	TotalAmount    int64      `json:"total_amount" gorm:"column:total_amount;not null"`
	Status         string     `json:"status" gorm:"column:status;type:varchar(50);default:'unpaid'"`
	DueDate        time.Time  `json:"due_date" gorm:"column:due_date;not null"`
	PeriodStart    time.Time  `json:"period_start" gorm:"column:period_start;not null"`
	PeriodEnd      time.Time  `json:"period_end" gorm:"column:period_end;not null"`
	Notes          string     `json:"notes" gorm:"column:notes;type:text"`
	PaidAt         *time.Time `json:"paid_at" gorm:"column:paid_at"`
	CreatedAt      time.Time  `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt      time.Time  `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

type InvoiceFilter struct {
	IDs             []string `json:"ids"`
	TenantIDs       []string `json:"tenant_ids"`
	CustomerIDs     []string `json:"customer_ids"`
	SubscriptionIDs []string `json:"subscription_ids"`
	Statuses        []string `json:"statuses"`
	InvoiceNumbers  []string `json:"invoice_numbers"`
}

func (f InvoiceFilter) IsEmpty() bool {
	return len(f.IDs) == 0 && len(f.TenantIDs) == 0 && len(f.CustomerIDs) == 0 &&
		len(f.SubscriptionIDs) == 0 && len(f.Statuses) == 0 && len(f.InvoiceNumbers) == 0
}
