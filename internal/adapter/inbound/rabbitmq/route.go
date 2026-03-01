package rabbitmq_inbound_adapter

import (
	"context"
	"os"

	inbound_port "go-template/internal/port/inbound"
	"go-template/utils/log"
	"go-template/utils/rabbitmq"
)

func InitRoute(
	ctx context.Context,
	args []string,
	port inbound_port.MessagePort,
) {
	if len(args) > 2 {
		switch args[2] {
		case "upsert_client":
			log.WithContext(ctx).Info("message subscribe upsert client started")
			done := make(chan struct{})
			go func() {
				// TODO: Use proper message type for admin user
				err := rabbitmq.Subscriber(
					"admin_user.upsert",
					rabbitmq.KindFanOut,
					os.Getenv("UPSERT_CLIENT_MESSAGE_SUBSCRIBE"),
					"",
					func(msg []byte) bool {
						return port.Client().Upsert(msg)
					},
				)
				if err != nil {
					log.WithContext(ctx).Error("failed to subscribe to message", err)
				}
				close(done)
			}()
			<-done
		case "mikrotik_sync":
			log.WithContext(ctx).Info("message subscribe mikrotik sync started")
			done := make(chan struct{})
			go func() {
				// Subscribe to PPPoE sync messages
				err := rabbitmq.Subscriber(
					"mikrotik.sync.pppoe",
					rabbitmq.KindTopic,
					"mikrotik.sync",
					"mikrotik.sync.pppoe",
					func(msg []byte) bool {
						return port.MikrotikSync().ProcessPPPoESync(msg)
					},
				)
				if err != nil {
					log.WithContext(ctx).Error("failed to subscribe to PPPoE sync", err)
				}
				close(done)
			}()
			
			done2 := make(chan struct{})
			go func() {
				// Subscribe to Hotspot sync messages
				err := rabbitmq.Subscriber(
					"mikrotik.sync.hotspot",
					rabbitmq.KindTopic,
					"mikrotik.sync",
					"mikrotik.sync.hotspot",
					func(msg []byte) bool {
						return port.MikrotikSync().ProcessHotspotSync(msg)
					},
				)
				if err != nil {
					log.WithContext(ctx).Error("failed to subscribe to Hotspot sync", err)
				}
				close(done2)
			}()
			
			<-done
			<-done2
		default:
			log.WithContext(ctx).Info("message subscribe not found")
		}
	} else {
		log.WithContext(ctx).Info("message subscribe not found")
	}
}
