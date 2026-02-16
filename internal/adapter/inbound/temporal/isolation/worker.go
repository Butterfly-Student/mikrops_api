package isolation_workflow

import (
	"context"

	"go-template/internal/domain"
	"go-template/utils/activity"
	"go-template/utils/log"
	"go-template/utils/temporal"
	"go.temporal.io/sdk/worker"
)

// StartIsolationWorker starts worker for isolation workflows
func StartIsolationWorker(ctx context.Context, domain domain.Domain) {
	activityCtx := activity.NewContext("isolation-worker")

	w, err := temporal.NewWorker(activityCtx, "IsolationTaskQueue")
	if err != nil {
		log.WithContext(activityCtx).Error("Unable to create isolation worker", err)
		return
	}

	activities := NewActivities(domain)

	// Register workflows
	w.RegisterWorkflow(IsolateExpiredCustomersWorkflow)
	w.RegisterWorkflow(ReactivateCustomerWorkflow)

	// Register activities
	w.RegisterActivity(activities.FindExpiredCustomersActivity)
	w.RegisterActivity(activities.CheckAutoIsolateActivity)
	w.RegisterActivity(activities.IsolateCustomerActivity)
	w.RegisterActivity(activities.SendIsolationNotificationActivity)
	w.RegisterActivity(activities.GetCustomerDetailsActivity)
	w.RegisterActivity(activities.GetOriginalProfileActivity)
	w.RegisterActivity(activities.ReactivateOnMikrotikActivity)
	w.RegisterActivity(activities.UpdateCustomerStatusActivity)
	w.RegisterActivity(activities.UpdateExpiryDateActivity)
	w.RegisterActivity(activities.SendReactivationNotificationActivity)

	// Start worker
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.WithContext(activityCtx).Error("Unable to start isolation worker", err)
	}
}
