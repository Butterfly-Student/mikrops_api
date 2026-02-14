package xendit

import (
	"time"
)

type Environment string

const (
	EnvironmentDevelopment Environment = "development"
	EnvironmentProduction  Environment = "production"
)

type CreateInvoiceRequest struct {
	ExternalID                     string            `json:"external_id"`
	Amount                         float64           `json:"amount"`
	InvoiceDuration                int               `json:"invoice_duration"`
	Description                    string            `json:"description"`
	Customer                       CustomerDetail    `json:"customer"`
	CustomerNotificationPreference string            `json:"customer_notification_preference"`
	PaymentMethods                 []string          `json:"payment_methods"`
	ShouldSendEmail                bool              `json:"should_send_email"`
	ShouldSendSMS                  bool              `json:"should_send_sms"`
	SuccessRedirectURL             string            `json:"success_redirect_url"`
	FailureRedirectURL             string            `json:"failure_redirect_url"`
	Meta                           map[string]string `json:"meta"`
}

type CustomerDetail struct {
	GivenNames   string `json:"given_names"`
	Email        string `json:"email"`
	MobileNumber string `json:"mobile_number,omitempty"`
}

type CreateInvoiceResponse struct {
	ID                     string            `json:"id"`
	ExternalID             string            `json:"external_id"`
	UserID                 string            `json:"user_id"`
	Status                 string            `json:"status"`
	MerchantName           string            `json:"merchant_name"`
	MerchantProfilePicture string            `json:"merchant_profile_picture"`
	Amount                 float64           `json:"amount"`
	InvoiceURL             string            `json:"invoice_url"`
	ExpiryDate             *time.Time        `json:"expiry_date"`
	BankCode               *string           `json:"bank_code"`
	PaymentChannels        []PaymentChannel  `json:"available_banks"`
	Description            string            `json:"description"`
	Created                *time.Time        `json:"created"`
	Updated                *time.Time        `json:"updated"`
	Meta                   map[string]string `json:"meta"`
}

type PaymentChannel struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type WebhookData struct {
	ExternalID         string            `json:"external_id"`
	Status             string            `json:"status"`
	PaymentMethod      string            `json:"payment_method"`
	Amount             float64           `json:"amount"`
	PaidAt             *time.Time        `json:"paid_at"`
	PaidAmount         float64           `json:"paid_amount"`
	BankCode           *string           `json:"bank_code"`
	AdjustmentAmount   float64           `json:"adjustment_amount"`
	RevenueAmount      float64           `json:"revenue_amount"`
	FeesPaidAmount     float64           `json:"fees_paid_amount"`
	Description        string            `json:"description"`
	Currency           string            `json:"currency"`
	PaymentID          string            `json:"payment_id"`
	PaymentChannel     string            `json:"payment_channel"`
	PaymentDestination string            `json:"payment_destination"`
	IsHighPriority     bool              `json:"is_high_priority"`
	Meta               map[string]string `json:"meta"`
}

type CreateVAPaymentRequest struct {
	ExternalID           string     `json:"external_id"`
	BankCode             string     `json:"bank_code"`
	Name                 string     `json:"name"`
	ExpectedAmount       float64    `json:"expected_amount"`
	ExpirationDate       *time.Time `json:"expiration_date"`
	IsSingleUse          bool       `json:"is_single_use"`
	VirtualAccountNumber string     `json:"virtual_account_number"`
	SuggestedAmount      *float64   `json:"suggested_amount"`
	Description          string     `json:"description"`
}

type CreateVAPaymentResponse struct {
	ID              string        `json:"id"`
	ExternalID      string        `json:"external_id"`
	OwnerID         string        `json:"owner_id"`
	BankCode        string        `json:"bank_code"`
	MerchantCode    string        `json:"merchant_code"`
	Name            string        `json:"name"`
	AccountNumber   string        `json:"account_number"`
	IsClosed        bool          `json:"is_closed"`
	ExpectedAmount  float64       `json:"expected_amount"`
	IsSingleUse     bool          `json:"is_single_use"`
	Status          string        `json:"status"`
	ExpirationDate  *time.Time    `json:"expiration_date"`
	SuggestedAmount *float64      `json:"suggested_amount"`
	Description     string        `json:"description"`
	Payment         PaymentDetail `json:"payment"`
}

type PaymentDetail struct {
	ID          string      `json:"id"`
	Amount      float64     `json:"amount"`
	Status      string      `json:"status"`
	PaidAt      *time.Time  `json:"paid_at"`
	PaidAmount  float64     `json:"paid_amount"`
	BankAccount BankAccount `json:"bank_account"`
}

type BankAccount struct {
	BankCode       string `json:"bank_code"`
	AccountNumber  string `json:"account_number"`
	AccountHolder  string `json:"account_holder"`
	PaymentChannel string `json:"payment_channel"`
}

type ErrorResponse struct {
	ErrorCode string `json:"error_code"`
	Message   string `json:"message"`
}
