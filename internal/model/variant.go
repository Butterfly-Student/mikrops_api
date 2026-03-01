package model

import "time"

// Variant is a legacy placeholder from the project template.
// It is kept to avoid breaking existing references but is no longer used in the ISP domain.
// TODO: remove once all references are cleaned up.
type Variant struct {
	ID        int       `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type VariantInput struct {
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type VariantFilter struct {
	IDs []int `json:"ids"`
}

func VariantPrepare(v *VariantInput) {
	v.CreatedAt = time.Now()
	v.UpdatedAt = time.Now()
}

func (c VariantFilter) IsEmpty() bool {
	return len(c.IDs) == 0
}
