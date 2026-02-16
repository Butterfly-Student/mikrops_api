package inbound_port

type WorkflowPort interface {
	Client() ClientWorkflowPort
	Billing() BillingWorkflowPort
	Isolation() IsolationWorkflowPort
}
