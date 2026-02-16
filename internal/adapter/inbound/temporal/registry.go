package temporal_inbound_adapter

import (
	billing_temporal_inbound_adapter "go-template/internal/adapter/inbound/temporal/billing"
	client_temporal_inbound_adapter "go-template/internal/adapter/inbound/temporal/client"
	isolation_temporal_inbound_adapter "go-template/internal/adapter/inbound/temporal/isolation"
	"go-template/internal/domain"
	inbound_port "go-template/internal/port/inbound"
)

type adapter struct {
	domain domain.Domain
}

func NewAdapter(
	domain domain.Domain,
) inbound_port.WorkflowPort {
	return &adapter{
		domain: domain,
	}
}

func (a *adapter) Client() inbound_port.ClientWorkflowPort {
	return client_temporal_inbound_adapter.NewClientAdapter(a.domain)
}

func (a *adapter) Billing() inbound_port.BillingWorkflowPort {
	return billing_temporal_inbound_adapter.NewBillingAdapter(a.domain)
}

func (a *adapter) Isolation() inbound_port.IsolationWorkflowPort {
	return isolation_temporal_inbound_adapter.NewIsolationAdapter(a.domain)
}
