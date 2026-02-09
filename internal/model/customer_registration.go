package model

import "time"

const (
	RegistrationStatusPending  = "pending"
	RegistrationStatusApproved = "approved"
	RegistrationStatusRejected = "rejected"
)

type CustomerRegistration struct {
	ID string `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	CustomerRegistrationInput
	Customer        *Customer        `json:"customer,omitempty" gorm:"foreignKey:CustomerID"`
	InternetPackage *InternetPackage `json:"internet_package,omitempty" gorm:"foreignKey:RequestedPackageID"`
	Nas             *Nas             `json:"nas,omitempty" gorm:"foreignKey:RequestedNasID"`
	ApprovedByStaff *Staff           `json:"approved_by_staff,omitempty" gorm:"foreignKey:ApprovedBy"`
}

func (CustomerRegistration) TableName() string {
	return "customer_registrations"
}

type CustomerRegistrationInput struct {
	TenantID           string     `json:"tenant_id" gorm:"column:tenant_id;type:uuid;not null;index"`
	CustomerID         *string    `json:"customer_id" gorm:"column:customer_id;type:uuid"`
	FullName           string     `json:"full_name" gorm:"column:full_name;type:varchar(255);not null"`
	Email              string     `json:"email" gorm:"column:email;type:varchar(255)"`
	Phone              string     `json:"phone" gorm:"column:phone;type:varchar(50)"`
	Address            string     `json:"address" gorm:"column:address;type:text"`
	RequestedPackageID string     `json:"requested_package_id" gorm:"column:requested_package_id;type:uuid;not null"`
	RequestedNasID     *string    `json:"requested_nas_id" gorm:"column:requested_nas_id;type:uuid"`
	Status             string     `json:"status" gorm:"column:status;type:varchar(50);default:'pending'"`
	RejectionReason    string     `json:"rejection_reason" gorm:"column:rejection_reason;type:text"`
	ApprovedBy         *string    `json:"approved_by" gorm:"column:approved_by;type:uuid"`
	ApprovedAt         *time.Time `json:"approved_at" gorm:"column:approved_at"`
	CreatedAt          time.Time  `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt          time.Time  `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

type CustomerRegistrationFilter struct {
	IDs       []string `json:"ids"`
	TenantIDs []string `json:"tenant_ids"`
	Statuses  []string `json:"statuses"`
	WithPackage  bool  `json:"-"`
	WithNas      bool  `json:"-"`
	WithCustomer bool  `json:"-"`
}

func (f CustomerRegistrationFilter) IsEmpty() bool {
	return len(f.IDs) == 0 && len(f.TenantIDs) == 0 && len(f.Statuses) == 0
}
