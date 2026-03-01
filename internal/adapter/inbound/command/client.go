package command_inbound_adapter

import (
	"context"

	"go-template/internal/domain"
	"go-template/internal/model"
	inbound_port "go-template/internal/port/inbound"
	"go-template/utils/activity"
	"go-template/utils/log"
)

type clientAdapter struct {
	domain domain.Domain
}

func NewClientAdapter(
	domain domain.Domain,
) inbound_port.ClientCommandPort {
	return &clientAdapter{
		domain: domain,
	}
}

func (h *clientAdapter) PublishUpsert(name string) {
	ctx := activity.NewContext("command_client_publish_upsert")
	ctx = context.WithValue(ctx, activity.Payload, name)
	// TODO: Adjust payload structure for AdminUserInput
	payload := []model.AdminUserInput{{FullName: name}}

	err := h.domain.Client().PublishUpsert(ctx, payload)
	if err != nil {
		log.WithContext(ctx).Error("client publish upsert error", err)
	}
	log.WithContext(ctx).Info("client publish upsert success")
}

func (h *clientAdapter) StartUpsert(name string) {
	ctx := activity.NewContext("command_client_start_upsert")
	ctx = context.WithValue(ctx, activity.Payload, name)
	// TODO: Adjust payload structure for AdminUserInput
	payload := model.AdminUserInput{FullName: name}

	err := h.domain.Client().StartUpsert(ctx, payload)
	if err != nil {
		log.WithContext(ctx).Error("client start upsert error", err)
	}
	log.WithContext(ctx).Info("client start upsert success")
}
