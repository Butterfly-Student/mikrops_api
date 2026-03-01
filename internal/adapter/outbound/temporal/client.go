package temporal_outbound_adapter

import (
	"context"
	"os"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
	"go-template/utils/temporal"
)

type clientWorkflowAdapter struct{}

func NewClientWorkflowAdapter() outbound_port.ClientWorkflowPort {
	return &clientWorkflowAdapter{}
}

func (g *clientWorkflowAdapter) StartUpsert(input model.AdminUserInput) error {
	namespace := os.Getenv("WORKFLOW_NAMESPACE")
	// TODO: Define proper workflow name for admin user upsert
	_, err := temporal.ExecuteWorkflow(context.Background(), namespace, "AdminUserUpsertWorkflow", input)
	if err != nil {
		return err
	}

	return nil
}
