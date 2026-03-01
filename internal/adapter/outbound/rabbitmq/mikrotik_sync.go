package rabbitmq_outbound_adapter

import (
	"context"
	"encoding/json"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
	"go-template/utils/rabbitmq"
)

type mikrotikSyncAdapter struct{}

// NewMikrotikSyncAdapter creates a new RabbitMQ adapter for MikroTik sync messages
func NewMikrotikSyncAdapter() outbound_port.MikrotikSyncMessagePort {
	return &mikrotikSyncAdapter{}
}

// PublishSyncMessage publishes a MikroTik sync message to RabbitMQ
func (a *mikrotikSyncAdapter) PublishSyncMessage(message model.MikrotikSyncMessage) error {
	// Marshal message to JSON
	payload, err := json.Marshal(message)
	if err != nil {
		return err
	}

	// Determine routing key based on service type
	routingKey := "mikrotik.sync.pppoe"
	switch message.ServiceType {
	case model.MikrotikServiceTypeHotspot:
		routingKey = "mikrotik.sync.hotspot"
	case model.MikrotikServiceTypeStaticIP:
		routingKey = "mikrotik.sync.static_ip"
	case model.MikrotikServiceTypeVPN:
		routingKey = "mikrotik.sync.vpn"
	}

	// Publish to RabbitMQ using topic exchange
	return rabbitmq.Publish(context.Background(), "mikrotik.sync", rabbitmq.KindTopic, routingKey, payload)
}
