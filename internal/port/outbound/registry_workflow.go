package outbound_port

import inbound_port "go-template/internal/port/inbound"

//go:generate mockgen -source=registry_workflow.go -destination=./../../../tests/mocks/port/mock_registry_workflow.go
type WorkflowPort interface {
	Client() ClientWorkflowPort
	Billing() inbound_port.BillingWorkflowPort
	Isolation() inbound_port.IsolationWorkflowPort
}
