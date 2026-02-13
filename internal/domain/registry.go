package domain

import (
	"go-template/internal/domain/auth"
	"go-template/internal/domain/client"
	"go-template/internal/domain/pppoe"
	"go-template/internal/domain/queue"
	"go-template/internal/domain/user"
	outbound_port "go-template/internal/port/outbound"

	"github.com/casbin/casbin/v2"
)

type Domain interface {
	Client() client.ClientDomain
	Auth() auth.AuthDomain
	User() user.UserDomain
	Pppoe() pppoe.PppoeDomain
	Queue() queue.QueueDomain
}

type domain struct {
	databasePort outbound_port.DatabasePort
	messagePort  outbound_port.MessagePort
	cachePort    outbound_port.CachePort
	workflowPort outbound_port.WorkflowPort
	mikrotikPort outbound_port.MikrotikPort
	enforcer     *casbin.Enforcer
}

func NewDomain(
	databasePort outbound_port.DatabasePort,
	messagePort outbound_port.MessagePort,
	cachePort outbound_port.CachePort,
	workflowPort outbound_port.WorkflowPort,
	mikrotikPort outbound_port.MikrotikPort,
	enforcer *casbin.Enforcer,
) Domain {
	return &domain{
		databasePort: databasePort,
		messagePort:  messagePort,
		cachePort:    cachePort,
		workflowPort: workflowPort,
		mikrotikPort: mikrotikPort,
		enforcer:     enforcer,
	}
}

func (d *domain) Client() client.ClientDomain {
	return client.NewClientDomain(d.databasePort, d.messagePort, d.cachePort, d.workflowPort)
}

func (d *domain) Auth() auth.AuthDomain {
	return auth.NewAuthDomain(d.databasePort, d.enforcer)
}

func (d *domain) User() user.UserDomain {
	return user.NewUserDomain(d.databasePort)
}

func (d *domain) Pppoe() pppoe.PppoeDomain {
	return pppoe.NewPppoeDomain(d.databasePort, d.cachePort, d.mikrotikPort)
}

func (d *domain) Queue() queue.QueueDomain {
	return queue.NewQueueDomain(d.databasePort, d.cachePort, d.mikrotikPort)
}
