package rabbitmq_outbound_adapter

import (
	"context"

	"mikrops/internal/model"
	outbound_port "mikrops/internal/port/outbound"
	"mikrops/utils/rabbitmq"
)

type clientAdapter struct{}

func NewClientAdapter() outbound_port.ClientMessagePort {
	return &clientAdapter{}
}

func (adapter *clientAdapter) PublishUpsert(datas []model.ClientInput) error {
	err := rabbitmq.Publish(context.Background(), model.UpsertClientMessage, rabbitmq.KindFanOut, "", datas)
	if err != nil {
		return err
	}

	return nil
}
