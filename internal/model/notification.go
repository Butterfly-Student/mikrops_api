package model

import (
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Type        string     `json:"type" gorm:"not null;type:varchar(50);index" validate:"required,oneof=email whatsapp sms"`
	Recipient   string     `json:"recipient" gorm:"not null;type:varchar(255)" validate:"required"`
	Subject     *string    `json:"subject" gorm:"type:varchar(500)"`
	Content     string     `json:"content" gorm:"type:text;not null" validate:"required"`
	CustomerID  *uuid.UUID `json:"customer_id" gorm:"type:uuid;index"`
	InvoiceID   *uuid.UUID `json:"invoice_id" gorm:"type:uuid;index"`
	PaymentID   *uuid.UUID `json:"payment_id" gorm:"type:uuid;index"`
	Status      string     `json:"status" gorm:"default:'pending';not null;type:varchar(20)" validate:"required,oneof=pending sent failed retrying"`
	ErrorCode   *string    `json:"error_code" gorm:"type:varchar(100)"`
	ErrorMsg    *string    `json:"error_msg" gorm:"type:text"`
	RetryCount  int        `json:"retry_count" gorm:"default:0"`
	ScheduledAt *time.Time `json:"scheduled_at" gorm:"type:timestamp"`
	SentAt      *time.Time `json:"sent_at" gorm:"type:timestamp"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   *time.Time `json:"-" gorm:"index"`
}

type NotificationType string

const (
	NotificationTypeEmail    NotificationType = "email"
	NotificationTypeWhatsApp NotificationType = "whatsapp"
	NotificationTypeSMS      NotificationType = "sms"
)

type NotificationTemplate struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name      string     `json:"name" gorm:"not null;unique;type:varchar(100)" validate:"required"`
	Type      string     `json:"type" gorm:"not null;type:varchar(20);index" validate:"required,oneof=email whatsapp sms"`
	Subject   string     `json:"subject" gorm:"not null;type:varchar(500)"`
	Content   string     `json:"content" gorm:"not null;type:text"`
	IsActive  bool       `json:"is_active" gorm:"default:true"`
	Variables string     `json:"variables" gorm:"type:jsonb"` // JSON array of required variables
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt *time.Time `json:"-" gorm:"index"`
}

type NotificationTemplateInput struct {
	Name      string `json:"name" validate:"required"`
	Type      string `json:"type" validate:"required,oneof=email whatsapp sms"`
	Subject   string `json:"subject" validate:"required"`
	Content   string `json:"content" validate:"required"`
	IsActive  *bool  `json:"is_active"`
	Variables string `json:"variables"`
}

type NotificationInput struct {
	Type       string     `json:"type" validate:"required,oneof=email whatsapp sms"`
	Recipient  string     `json:"recipient" validate:"required,email"`
	Subject    *string    `json:"subject"`
	Content    string     `json:"content" validate:"required"`
	CustomerID *uuid.UUID `json:"customer_id"`
	InvoiceID  *uuid.UUID `json:"invoice_id"`
	PaymentID  *uuid.UUID `json:"payment_id"`
}

type NotificationFilter struct {
	IDs         []uuid.UUID `json:"ids"`
	Types       []string    `json:"types"`
	Status      []string    `json:"status"`
	Recipients  []string    `json:"recipients"`
	CustomerIDs []uuid.UUID `json:"customer_ids"`
	InvoiceIDs  []uuid.UUID `json:"invoice_ids"`
	PaymentIDs  []uuid.UUID `json:"payment_ids"`
	SentStart   *time.Time  `json:"sent_start"`
	SentEnd     *time.Time  `json:"sent_end"`
	Search      *string     `json:"search"`
}

func (f NotificationFilter) IsEmpty() bool {
	return len(f.IDs) == 0 && len(f.Types) == 0 &&
		len(f.Status) == 0 && len(f.Recipients) == 0 &&
		len(f.CustomerIDs) == 0 && len(f.InvoiceIDs) == 0 &&
		len(f.PaymentIDs) == 0 && f.SentStart == nil &&
		f.SentEnd == nil && f.Search == nil
}
