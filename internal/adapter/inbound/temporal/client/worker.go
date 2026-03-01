package client_temporal_inbound_adapter

import (
	"go.temporal.io/sdk/worker"

	"go-template/internal/domain"
	inbound_port "go-template/internal/port/inbound"
	"go-template/utils/activity"
	"go-template/utils/log"
	"go-template/utils/temporal"
)

type clientAdapter struct {
	domain domain.Domain
}

func NewClientAdapter(
	domain domain.Domain,
) inbound_port.ClientWorkflowPort {
	return &clientAdapter{
		domain: domain,
	}
}

func (a *clientAdapter) Upsert() {
	ctx := activity.NewContext("upsert_client_worker")

	// TODO: Use proper workflow name for admin user
	w, err := temporal.NewWorker(ctx, "AdminUserUpsertWorkflow")
	if err != nil {
		log.WithContext(ctx).Error("Unable to create worker", err)
		return
	}

	workflow := NewClientWorkflow(a.domain)

	w.RegisterWorkflow(workflow.UpsertClientWorkflow)
	w.RegisterActivity(a.domain.Client())

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.WithContext(ctx).Error("Unable to start worker", err)
		return
	}
}
