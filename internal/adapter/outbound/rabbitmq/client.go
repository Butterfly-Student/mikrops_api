package rabbitmq_outbound_adapter

import (
	"context"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
	"go-template/utils/rabbitmq"
)

type clientAdapter struct{}

func NewClientAdapter() outbound_port.ClientMessagePort {
	return &clientAdapter{}
}

func (adapter *clientAdapter) PublishUpsert(datas []model.AdminUserInput) error {
	// TODO: Define proper message type for admin user upsert
	err := rabbitmq.Publish(context.Background(), "admin_user.upsert", rabbitmq.KindFanOut, "", datas)
	if err != nil {
		return err
	}

	return nil
}
