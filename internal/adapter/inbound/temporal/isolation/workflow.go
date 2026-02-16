package isolation_workflow

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// IsolateExpiredCustomersWorkflow is a workflow that finds and isolates customers with expired subscriptions
func IsolateExpiredCustomersWorkflow(ctx workflow.Context, input IsolateExpiredCustomersInput) (IsolateExpiredCustomersResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting IsolateExpiredCustomersWorkflow")

	result := IsolateExpiredCustomersResult{
		TotalCustomers:    0,
		IsolatedCustomers: 0,
		FailedCustomers:   0,
		Errors:            []string{},
	}

	// Set activity options
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	// Step 1: Find expired customers
	var expiredCustomers FindExpiredCustomersResult
	err := workflow.ExecuteActivity(ctx, "FindExpiredCustomersActivity", FindExpiredCustomersInput{
		GracePeriodDays: input.GracePeriodDays,
	}).Get(ctx, &expiredCustomers)
	if err != nil {
		logger.Error("Failed to find expired customers", "error", err)
		return result, err
	}

	result.TotalCustomers = len(expiredCustomers.CustomerIDs)
	logger.Info("Found expired customers", "count", result.TotalCustomers)

	if result.TotalCustomers == 0 {
		logger.Info("No expired customers found, workflow complete")
		return result, nil
	}

	// Step 2: Process each customer
	for _, customerID := range expiredCustomers.CustomerIDs {
		// Check if auto-isolate is enabled for this customer
		var autoIsolateCheck CheckAutoIsolateResult
		err := workflow.ExecuteActivity(ctx, "CheckAutoIsolateActivity", CheckAutoIsolateInput{
			CustomerID: customerID,
		}).Get(ctx, &autoIsolateCheck)
		if err != nil {
			logger.Warn("Failed to check auto-isolate flag", "customerID", customerID, "error", err)
			result.FailedCustomers++
			result.Errors = append(result.Errors, customerID+": "+err.Error())
			continue
		}

		if !autoIsolateCheck.AutoIsolate {
			logger.Info("Skipping customer (auto-isolate disabled)", "customerID", customerID)
			continue
		}

		// Isolate the customer
		var isolateResult IsolateCustomerResult
		err = workflow.ExecuteActivity(ctx, "IsolateCustomerActivity", IsolateCustomerInput{
			CustomerID: customerID,
		}).Get(ctx, &isolateResult)
		if err != nil {
			logger.Error("Failed to isolate customer", "customerID", customerID, "error", err)
			result.FailedCustomers++
			result.Errors = append(result.Errors, customerID+": "+err.Error())
			continue
		}

		// Send isolation notification
		err = workflow.ExecuteActivity(ctx, "SendIsolationNotificationActivity", SendIsolationNotificationInput{
			CustomerID: customerID,
		}).Get(ctx, nil)
		if err != nil {
			logger.Warn("Failed to send isolation notification", "customerID", customerID, "error", err)
			// Don't fail the workflow if notification fails
		}

		result.IsolatedCustomers++
		logger.Info("Successfully isolated customer", "customerID", customerID)
	}

	logger.Info("IsolateExpiredCustomersWorkflow completed",
		"total", result.TotalCustomers,
		"isolated", result.IsolatedCustomers,
		"failed", result.FailedCustomers)

	return result, nil
}

// ReactivateCustomerWorkflow is a workflow that reactivates a customer after payment
func ReactivateCustomerWorkflow(ctx workflow.Context, input ReactivateCustomerInput) (ReactivateCustomerResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting ReactivateCustomerWorkflow", "customerID", input.CustomerID)

	result := ReactivateCustomerResult{
		Success: false,
	}

	// Set activity options
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	// Step 1: Get customer details
	var customerDetails GetCustomerDetailsResult
	err := workflow.ExecuteActivity(ctx, "GetCustomerDetailsActivity", GetCustomerDetailsInput{
		CustomerID: input.CustomerID,
	}).Get(ctx, &customerDetails)
	if err != nil {
		logger.Error("Failed to get customer details", "error", err)
		return result, err
	}

	// Step 2: Get original profile
	var originalProfile GetOriginalProfileResult
	err = workflow.ExecuteActivity(ctx, "GetOriginalProfileActivity", GetOriginalProfileInput{
		CustomerID: input.CustomerID,
	}).Get(ctx, &originalProfile)
	if err != nil {
		logger.Error("Failed to get original profile", "error", err)
		return result, err
	}

	// Step 3: Reactivate customer on MikroTik
	var reactivateResult ReactivateOnMikrotikResult
	err = workflow.ExecuteActivity(ctx, "ReactivateOnMikrotikActivity", ReactivateOnMikrotikInput{
		CustomerID: input.CustomerID,
		ProfileID:  originalProfile.ProfileID,
	}).Get(ctx, &reactivateResult)
	if err != nil {
		logger.Error("Failed to reactivate customer on MikroTik", "error", err)
		return result, err
	}

	// Step 3: Update customer status in database
	err = workflow.ExecuteActivity(ctx, "UpdateCustomerStatusActivity", UpdateCustomerStatusInput{
		CustomerID: input.CustomerID,
		Status:     "active",
	}).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to update customer status", "error", err)
		return result, err
	}

	// Step 4: Update expiry date
	err = workflow.ExecuteActivity(ctx, "UpdateExpiryDateActivity", UpdateExpiryDateInput{
		CustomerID: input.CustomerID,
		InvoiceID:  input.InvoiceID,
	}).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to update expiry date", "error", err)
		// Don't fail the workflow if expiry date update fails
	}

	// Step 5: Send reactivation notification
	err = workflow.ExecuteActivity(ctx, "SendReactivationNotificationActivity", SendReactivationNotificationInput{
		CustomerID: input.CustomerID,
	}).Get(ctx, nil)
	if err != nil {
		logger.Warn("Failed to send reactivation notification", "error", err)
		// Don't fail the workflow if notification fails
	}

	result.Success = true
	logger.Info("ReactivateCustomerWorkflow completed successfully", "customerID", input.CustomerID)

	return result, nil
}

// Workflow input/output types

type IsolateExpiredCustomersInput struct {
	GracePeriodDays int
}

type IsolateExpiredCustomersResult struct {
	TotalCustomers    int
	IsolatedCustomers int
	FailedCustomers   int
	Errors            []string
}

type ReactivateCustomerInput struct {
	CustomerID string
	InvoiceID  string
}

type ReactivateCustomerResult struct {
	Success bool
}

// Activity input/output types

// FindExpiredCustomers activity
type FindExpiredCustomersInput struct {
	GracePeriodDays int
}

type FindExpiredCustomersResult struct {
	CustomerIDs []string
}

// CheckAutoIsolate activity
type CheckAutoIsolateInput struct {
	CustomerID string
}

type CheckAutoIsolateResult struct {
	AutoIsolate bool
}

// IsolateCustomer activity
type IsolateCustomerInput struct {
	CustomerID string
}

type IsolateCustomerResult struct {
	Success bool
}

// SendIsolationNotification activity
type SendIsolationNotificationInput struct {
	CustomerID string
}

// GetCustomerDetails activity
type GetCustomerDetailsInput struct {
	CustomerID string
}

type GetCustomerDetailsResult struct {
	CustomerCode string
	FullName     string
	Email        string
	Phone        string
}

// ReactivateOnMikrotik activity
type ReactivateOnMikrotikInput struct {
	CustomerID string
	ProfileID  string
}

type ReactivateOnMikrotikResult struct {
	Success bool
}

// UpdateCustomerStatus activity
type UpdateCustomerStatusInput struct {
	CustomerID string
	Status     string
}

// UpdateExpiryDate activity
type UpdateExpiryDateInput struct {
	CustomerID string
	InvoiceID  string
}

// SendReactivationNotification activity
type SendReactivationNotificationInput struct {
	CustomerID string
}
