package client_temporal_inbound_adapter

import (
	"time"

	"go.temporal.io/sdk/workflow"

	"go-template/internal/domain"
	"go-template/internal/model"
)

type ClientWorkflow interface {
	UpsertClientWorkflow(ctx workflow.Context, input model.AdminUserInput) (string, error)
}

type clientWorkflow struct {
	domain domain.Domain
}

func NewClientWorkflow(
	domain domain.Domain,
) ClientWorkflow {
	return &clientWorkflow{
		domain: domain,
	}
}

func (g *clientWorkflow) UpsertClientWorkflow(ctx workflow.Context, input model.AdminUserInput) (string, error) {
	logger := workflow.GetLogger(ctx)
	workflowInfo := workflow.GetInfo(ctx)

	logger.Info("Workflow started", "WorkflowID", workflowInfo.WorkflowExecution.ID)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var results []model.AdminUser
	err := workflow.ExecuteActivity(
		ctx,
		g.domain.Client().Upsert,
		[]model.AdminUserInput{input},
	).Get(ctx, &results)
	if err != nil {
		logger.Error("UpsertClient activity failed", "Error", err)
		return "Failed to upsert client", err
	}

	var email string
	if len(results) > 0 {
		email = results[0].Email
	}

	successMessage := "User email: " + email
	logger.Info(successMessage, "WorkflowID", workflowInfo.WorkflowExecution.ID)

	return successMessage, nil
}
