package command_inbound_adapter

import (
	"context"

	"mikrops/internal/domain"
	inbound_port "mikrops/internal/port/inbound"
	"mikrops/utils/log"
)

func InitRoute(
	ctx context.Context,
	args []string,
	port inbound_port.CommandPort,
	domainRegistry domain.Domain,
) {
	if len(args) > 2 {
		switch args[1] {
		case "publish_upsert_client":
			name := args[2]
			port.Client().PublishUpsert(name)
		case "start_upsert_client":
			name := args[2]
			port.Client().StartUpsert(name)
		case "daily_cutoff":
			log.WithContext(ctx).Info("Running daily cutoff command")
			cutoffCmd := NewCutoffCommand(domainRegistry)
			err := cutoffCmd.Run()
			if err != nil {
				log.WithContext(ctx).Errorf("Daily cutoff command failed: %v", err)
			}
		default:
			log.WithContext(ctx).Info("command not found")
		}
	} else {
		log.WithContext(ctx).Info("command not found")
	}
}
