package temporal_outbound_adapter

import (
	"context"
	billing_adapter "go-template/internal/adapter/inbound/temporal/billing"
	isolation_adapter "go-template/internal/adapter/inbound/temporal/isolation"
	"go-template/internal/domain"
	inbound_port "go-template/internal/port/inbound"
	outbound_port "go-template/internal/port/outbound"
)

type adapter struct {
	domain           domain.Domain
	billingAdapter   inbound_port.BillingWorkflowPort
	isolationAdapter inbound_port.IsolationWorkflowPort
}

func NewAdapter(d domain.Domain) outbound_port.WorkflowPort {
	billingAdapter := billing_adapter.NewBillingAdapter(d)
	isolationAdapter := isolation_adapter.NewIsolationAdapter(d)
	return &adapter{
		domain:           d,
		billingAdapter:   billingAdapter,
		isolationAdapter: isolationAdapter,
	}
}

func (a *adapter) Client() outbound_port.ClientWorkflowPort {
	return nil
}

func (a *adapter) Billing() inbound_port.BillingWorkflowPort {
	return a.billingAdapter
}

func (a *adapter) Isolation() inbound_port.IsolationWorkflowPort {
	return a.isolationAdapter
}

func (a *adapter) StartBillingScheduler(ctx context.Context) {
	a.billingAdapter.StartScheduler(ctx)
}

func (a *adapter) StartIsolationScheduler(ctx context.Context) {
	a.isolationAdapter.StartScheduler(ctx)
}
