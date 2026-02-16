package model

import (
	"time"

	"github.com/google/uuid"
)

type ActivityLog struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID      *uuid.UUID `json:"user_id" gorm:"type:uuid;index"`
	Action      string     `json:"action" gorm:"not null;type:varchar(50);index" validate:"required,oneof=create update delete login logout isolate reactivate send_notification sync_to_mikrotik generate_invoice apply_payment"`
	EntityType  string     `json:"entity_type" gorm:"not null;type:varchar(50);index" validate:"required,oneof=customer invoice payment bandwidth_profile mikrotik_router user cash_transaction cash_category"`
	EntityID    *uuid.UUID `json:"entity_id" gorm:"type:uuid;index"`
	Description string     `json:"description" gorm:"not null;type:text"`
	OldValues   string     `json:"old_values" gorm:"type:jsonb"`
	NewValues   string     `json:"new_values" gorm:"type:jsonb"`
	IPAddress   string     `json:"ip_address" gorm:"type:varchar(45)"`
	UserAgent   string     `json:"user_agent" gorm:"type:text"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime;index"`
}

type ActivityLogInput struct {
	UserID      *uuid.UUID `json:"user_id"`
	Action      string     `json:"action" validate:"required,oneof=create update delete login logout isolate reactivate send_notification sync_to_mikrotik generate_invoice apply_payment"`
	EntityType  string     `json:"entity_type" validate:"required,oneof=customer invoice payment bandwidth_profile mikrotik_router user cash_transaction cash_category"`
	EntityID    *uuid.UUID `json:"entity_id"`
	Description string     `json:"description" validate:"required"`
	OldValues   string     `json:"old_values"`
	NewValues   string     `json:"new_values"`
	IPAddress   string     `json:"ip_address"`
	UserAgent   string     `json:"user_agent"`
}

type ActivityLogFilter struct {
	IDs            []uuid.UUID `json:"ids"`
	UserIDs        []uuid.UUID `json:"user_ids"`
	Actions        []string    `json:"actions"`
	EntityTypes    []string    `json:"entity_types"`
	EntityIDs      []uuid.UUID `json:"entity_ids"`
	CreatedAtStart *time.Time  `json:"created_at_start"`
	CreatedAtEnd   *time.Time  `json:"created_at_end"`
	Search         *string     `json:"search"`
	Limit          int         `json:"limit"`
	Offset         int         `json:"offset"`
}

func (f ActivityLogFilter) IsEmpty() bool {
	return len(f.IDs) == 0 && len(f.UserIDs) == 0 &&
		len(f.Actions) == 0 && len(f.EntityTypes) == 0 &&
		len(f.EntityIDs) == 0 && f.CreatedAtStart == nil &&
		f.CreatedAtEnd == nil && f.Search == nil
}
