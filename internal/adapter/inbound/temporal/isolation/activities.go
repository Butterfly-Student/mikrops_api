package isolation_workflow

import (
	"context"
	"fmt"
	"time"

	"go-template/internal/domain"
	"go-template/internal/model"
	"go-template/utils/log"
)

// Activities struct holds dependencies for workflow activities
type Activities struct {
	domain domain.Domain
}

// NewActivities creates a new Activities instance
func NewActivities(domain domain.Domain) *Activities {
	return &Activities{
		domain: domain,
	}
}

// FindExpiredCustomersActivity finds all expired customers
func (a *Activities) FindExpiredCustomersActivity(ctx context.Context, input FindExpiredCustomersInput) (FindExpiredCustomersResult, error) {
	log.WithContext(ctx).Info(fmt.Sprintf("Finding expired customers with grace period: %d days", input.GracePeriodDays))

	// Find expired customers
	customers, err := a.domain.Customer().FindExpired(ctx)
	if err != nil {
		return FindExpiredCustomersResult{}, fmt.Errorf("failed to find expired customers: %w", err)
	}

	// Extract customer IDs
	customerIDs := make([]string, 0, len(customers))
	for _, customer := range customers {
		// Check if grace period has passed
		if customer.ExpiryDate != nil {
			gracePeriodEnd := customer.ExpiryDate.Add(time.Duration(input.GracePeriodDays) * 24 * time.Hour)
			if time.Now().After(gracePeriodEnd) {
				customerIDs = append(customerIDs, customer.ID.String())
			}
		}
	}

	log.WithContext(ctx).Info(fmt.Sprintf("Found %d expired customers past grace period", len(customerIDs)))

	return FindExpiredCustomersResult{
		CustomerIDs: customerIDs,
	}, nil
}

// CheckAutoIsolateActivity checks if customer has auto-isolate enabled
func (a *Activities) CheckAutoIsolateActivity(ctx context.Context, input CheckAutoIsolateInput) (CheckAutoIsolateResult, error) {
	log.WithContext(ctx).Info(fmt.Sprintf("Checking auto-isolate flag for customer: %s", input.CustomerID))

	customer, err := a.domain.Customer().GetCustomer(ctx, input.CustomerID)
	if err != nil {
		return CheckAutoIsolateResult{}, fmt.Errorf("failed to get customer: %w", err)
	}

	autoIsolate := true
	if customer.AutoIsolate != nil {
		autoIsolate = *customer.AutoIsolate
	}

	return CheckAutoIsolateResult{
		AutoIsolate: autoIsolate,
	}, nil
}

// IsolateCustomerActivity isolates a customer
func (a *Activities) IsolateCustomerActivity(ctx context.Context, input IsolateCustomerInput) (IsolateCustomerResult, error) {
	log.WithContext(ctx).Info(fmt.Sprintf("Isolating customer: %s", input.CustomerID))

	err := a.domain.Customer().IsolateCustomer(ctx, input.CustomerID)
	if err != nil {
		return IsolateCustomerResult{Success: false}, fmt.Errorf("failed to isolate customer: %w", err)
	}

	log.WithContext(ctx).Info(fmt.Sprintf("Successfully isolated customer: %s", input.CustomerID))

	return IsolateCustomerResult{Success: true}, nil
}

// SendIsolationNotificationActivity sends isolation notification to customer
func (a *Activities) SendIsolationNotificationActivity(ctx context.Context, input SendIsolationNotificationInput) error {
	log.WithContext(ctx).Info(fmt.Sprintf("Sending isolation notification to customer: %s", input.CustomerID))

	// Get customer details
	customer, err := a.domain.Customer().GetCustomer(ctx, input.CustomerID)
	if err != nil {
		return fmt.Errorf("failed to get customer: %w", err)
	}

	// Send notification via notification domain
	reason := fmt.Sprintf("Your internet service has been isolated due to unpaid invoice")
	err = a.domain.Notification().SendIsolationNotification(ctx, input.CustomerID, reason)
	if err != nil {
		log.WithContext(ctx).Warn(fmt.Sprintf("Failed to send isolation notification: %v", err))
		// Don't fail the workflow if notification fails
	} else {
		log.WithContext(ctx).Info(fmt.Sprintf("Isolation notification sent to %s (%s)", customer.FullName, customer.Phone))
	}

	return nil
}

// GetCustomerDetailsActivity gets customer details
func (a *Activities) GetCustomerDetailsActivity(ctx context.Context, input GetCustomerDetailsInput) (GetCustomerDetailsResult, error) {
	log.WithContext(ctx).Info(fmt.Sprintf("Getting customer details: %s", input.CustomerID))

	customer, err := a.domain.Customer().GetCustomer(ctx, input.CustomerID)
	if err != nil {
		return GetCustomerDetailsResult{}, fmt.Errorf("failed to get customer: %w", err)
	}

	email := ""
	if customer.Email != nil {
		email = *customer.Email
	}

	return GetCustomerDetailsResult{
		CustomerCode: customer.CustomerCode,
		FullName:     customer.FullName,
		Email:        email,
		Phone:        customer.Phone,
	}, nil
}

// GetOriginalProfileActivity gets the customer's original bandwidth profile before isolation
func (a *Activities) GetOriginalProfileActivity(ctx context.Context, input GetOriginalProfileInput) (GetOriginalProfileResult, error) {
	log.WithContext(ctx).Info(fmt.Sprintf("Getting original profile for customer: %s", input.CustomerID))

	customer, err := a.domain.Customer().GetCustomer(ctx, input.CustomerID)
	if err != nil {
		return GetOriginalProfileResult{}, fmt.Errorf("failed to get customer: %w", err)
	}

	if customer.Profile == nil {
		return GetOriginalProfileResult{}, fmt.Errorf("customer has no profile set")
	}

	profile, err := a.domain.BandwidthProfile().GetProfile(ctx, customer.Profile.ID.String())
	if err != nil {
		return GetOriginalProfileResult{}, fmt.Errorf("failed to get profile: %w", err)
	}

	log.WithContext(ctx).Info(fmt.Sprintf("Found original profile for customer %s: %s", customer.CustomerCode, profile.Name))

	return GetOriginalProfileResult{
		ProfileID:      profile.ID.String(),
		ProfileName:    profile.Name,
		PppProfileName: profile.PppProfileName,
		ProfileCode:    profile.ProfileCode,
	}, nil
}

// ReactivateOnMikrotikActivity reactivates customer on MikroTik
func (a *Activities) ReactivateOnMikrotikActivity(ctx context.Context, input ReactivateOnMikrotikInput) (ReactivateOnMikrotikResult, error) {
	log.WithContext(ctx).Info(fmt.Sprintf("Reactivating customer on MikroTik: %s", input.CustomerID))

	err := a.domain.Customer().ActivateCustomer(ctx, input.CustomerID)
	if err != nil {
		return ReactivateOnMikrotikResult{Success: false}, fmt.Errorf("failed to activate customer: %w", err)
	}

	log.WithContext(ctx).Info(fmt.Sprintf("Successfully reactivated customer: %s", input.CustomerID))

	return ReactivateOnMikrotikResult{Success: true}, nil
}

// UpdateCustomerStatusActivity updates customer status
func (a *Activities) UpdateCustomerStatusActivity(ctx context.Context, input UpdateCustomerStatusInput) error {
	log.WithContext(ctx).Info(fmt.Sprintf("Updating customer status: %s to %s", input.CustomerID, input.Status))

	customer, err := a.domain.Customer().GetCustomer(ctx, input.CustomerID)
	if err != nil {
		return fmt.Errorf("failed to get customer: %w", err)
	}

	customer.Status = input.Status

	// Update customer via domain method
	_, err = a.domain.Customer().UpdateCustomer(ctx, input.CustomerID, model.CustomerInput{
		CustomerCode: &customer.CustomerCode,
		FullName:     customer.FullName,
		Email:        customer.Email,
		Phone:        customer.Phone,
		Address:      customer.Address,
		Status:       &customer.Status,
	})
	if err != nil {
		return fmt.Errorf("failed to update customer: %w", err)
	}

	log.WithContext(ctx).Info(fmt.Sprintf("Customer status updated to: %s", input.Status))

	return nil
}

// UpdateExpiryDateActivity updates customer expiry date
func (a *Activities) UpdateExpiryDateActivity(ctx context.Context, input UpdateExpiryDateInput) error {
	log.WithContext(ctx).Info(fmt.Sprintf("Updating expiry date for customer: %s, invoice: %s", input.CustomerID, input.InvoiceID))

	// Get customer details
	customer, err := a.domain.Customer().GetCustomer(ctx, input.CustomerID)
	if err != nil {
		return fmt.Errorf("failed to get customer: %w", err)
	}

	// Calculate new expiry date based on billing cycle
	var newExpiryDate time.Time

	if customer.BillingCycle == nil || *customer.BillingCycle == "monthly" {
		// Add one month to current expiry date (or today if nil)
		currentExpiry := time.Now()
		if customer.ExpiryDate != nil {
			currentExpiry = *customer.ExpiryDate
		}
		newExpiryDate = currentExpiry.AddDate(0, 1, 0)
	} else if customer.BillingCycle != nil && *customer.BillingCycle == "quarterly" {
		// Add three months
		currentExpiry := time.Now()
		if customer.ExpiryDate != nil {
			currentExpiry = *customer.ExpiryDate
		}
		newExpiryDate = currentExpiry.AddDate(0, 3, 0)
	} else if customer.BillingCycle != nil && *customer.BillingCycle == "yearly" {
		// Add one year
		currentExpiry := time.Now()
		if customer.ExpiryDate != nil {
			currentExpiry = *customer.ExpiryDate
		}
		newExpiryDate = currentExpiry.AddDate(1, 0, 0)
	} else {
		// Default to one month
		currentExpiry := time.Now()
		if customer.ExpiryDate != nil {
			currentExpiry = *customer.ExpiryDate
		}
		newExpiryDate = currentExpiry.AddDate(0, 1, 0)
	}

	// Update customer via domain method
	_, err = a.domain.Customer().UpdateCustomer(ctx, input.CustomerID, model.CustomerInput{
		CustomerCode: &customer.CustomerCode,
		FullName:     customer.FullName,
		Email:        customer.Email,
		Phone:        customer.Phone,
		Address:      customer.Address,
		Status:       &customer.Status,
		ExpiryDate:   &newExpiryDate,
	})
	if err != nil {
		return fmt.Errorf("failed to update customer: %w", err)
	}

	log.WithContext(ctx).Info(fmt.Sprintf("Expiry date updated to: %s for customer %s", newExpiryDate.Format("2006-01-02"), customer.CustomerCode))

	return nil
}

// SendReactivationNotificationActivity sends reactivation notification to customer
func (a *Activities) SendReactivationNotificationActivity(ctx context.Context, input SendReactivationNotificationInput) error {
	log.WithContext(ctx).Info(fmt.Sprintf("Sending reactivation notification to customer: %s", input.CustomerID))

	// Get customer details
	customer, err := a.domain.Customer().GetCustomer(ctx, input.CustomerID)
	if err != nil {
		return fmt.Errorf("failed to get customer: %w", err)
	}

	// Send notification via notification domain
	err = a.domain.Notification().SendReactivationNotification(ctx, input.CustomerID, "Your internet service has been reactivated successfully. Thank you for your payment!")
	if err != nil {
		log.WithContext(ctx).Warn(fmt.Sprintf("Failed to send reactivation notification: %v", err))
		// Don't fail the workflow if notification fails
	} else {
		log.WithContext(ctx).Info(fmt.Sprintf("Reactivation notification sent to %s (%s)", customer.FullName, customer.Phone))
	}

	return nil
}

// GetOriginalProfileInput is input for getting original profile
type GetOriginalProfileInput struct {
	CustomerID string
}

// GetOriginalProfileResult is output for getting original profile
type GetOriginalProfileResult struct {
	ProfileID      string
	ProfileName    string
	PppProfileName string
	ProfileCode    string
}
