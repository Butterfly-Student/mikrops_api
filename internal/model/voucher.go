package model

import (
	"time"

	"github.com/google/uuid"
)

// ─── Voucher Batches ──────────────────────────────────────────────────────────

// VoucherBatch represents a batch of hotspot vouchers
type VoucherBatch struct {
	ID            uuid.UUID         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	BatchCode     string            `gorm:"size:50;uniqueIndex;not null" json:"batch_code" validate:"required,max=50"`
	PlanID        uuid.UUID         `gorm:"type:uuid;not null" json:"plan_id" validate:"required"`
	Plan          *BandwidthProfile `gorm:"foreignKey:PlanID;constraint:OnDelete:RESTRICT" json:"plan,omitempty"`
	RouterID      uuid.UUID         `gorm:"type:uuid;not null" json:"router_id" validate:"required"`
	Router        *MikrotikRouter   `gorm:"foreignKey:RouterID;constraint:OnDelete:RESTRICT" json:"router,omitempty"`
	TotalVouchers int               `gorm:"not null" json:"total_vouchers" validate:"required,min=1"`
	SoldVouchers  int               `gorm:"default:0" json:"sold_vouchers"`
	UsedVouchers  int               `gorm:"default:0" json:"used_vouchers"`
	PriceOverride *float64          `gorm:"type:decimal(12,2)" json:"price_override,omitempty"` // NULL = use plan price
	Notes         *string           `gorm:"type:text" json:"notes,omitempty"`
	CreatedBy     uuid.UUID         `gorm:"type:uuid;not null" json:"created_by"`
	Creator       *AdminUser        `gorm:"foreignKey:CreatedBy;constraint:OnDelete:RESTRICT" json:"creator,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`

	// Relations
	Vouchers []Voucher `gorm:"foreignKey:BatchID" json:"vouchers,omitempty"`
}

func (VoucherBatch) TableName() string { return "voucher_batches" }

// VoucherStatus represents the lifecycle state of a single voucher
type VoucherStatus string

const (
	VoucherStatusAvailable VoucherStatus = "available"
	VoucherStatusSold      VoucherStatus = "sold"
	VoucherStatusUsed      VoucherStatus = "used"
	VoucherStatusExpired   VoucherStatus = "expired"
)

// Voucher represents a single hotspot voucher
type Voucher struct {
	ID        uuid.UUID     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	BatchID   uuid.UUID     `gorm:"type:uuid;not null" json:"batch_id" validate:"required"`
	Batch     *VoucherBatch `gorm:"foreignKey:BatchID;constraint:OnDelete:RESTRICT" json:"batch,omitempty"`
	Code      string        `gorm:"size:30;uniqueIndex;not null" json:"code" validate:"required"`
	Username  string        `gorm:"size:100;not null" json:"username"` // MikroTik hotspot username
	Password  string        `gorm:"size:100;not null" json:"password"` // MikroTik hotspot password
	Status    VoucherStatus `gorm:"size:20;not null;default:available" json:"status"`
	SoldTo    *string       `gorm:"size:100" json:"sold_to,omitempty"`
	SoldAt    *time.Time    `json:"sold_at,omitempty"`
	UsedBy    *string       `gorm:"size:100" json:"used_by,omitempty"`
	UsedAt    *time.Time    `json:"used_at,omitempty"`
	ExpiredAt *time.Time    `json:"expired_at,omitempty"`
	MtSynced  bool          `gorm:"default:false" json:"mt_synced"`
}

func (Voucher) TableName() string { return "vouchers" }

// VoucherBatchInput for creating a new voucher batch
type VoucherBatchInput struct {
	BatchCode     string    `json:"batch_code" validate:"required,max=50"`
	PlanID        uuid.UUID `json:"plan_id" validate:"required"`
	RouterID      uuid.UUID `json:"router_id" validate:"required"`
	TotalVouchers int       `json:"total_vouchers" validate:"required,min=1"`
	PriceOverride *float64  `json:"price_override" validate:"omitempty,min=0"`
	Notes         *string   `json:"notes"`
}

// VoucherFilter for querying vouchers
type VoucherFilter struct {
	BatchID *uuid.UUID     `json:"batch_id"`
	Status  *VoucherStatus `json:"status"`
	Search  *string        `json:"search"` // code, sold_to, used_by
}
