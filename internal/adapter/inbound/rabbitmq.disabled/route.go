package rabbitmq_inbound_adapter

import (
	"context"
	"os"

	"mikrops/internal/domain"
	"mikrops/internal/model"
	inbound_port "mikrops/internal/port/inbound"
	"mikrops/utils/log"
	"mikrops/utils/rabbitmq"
)

func InitRoute(
	ctx context.Context,
	args []string,
	port inbound_port.MessagePort,
	domainRegistry domain.Domain,
) {
	if len(args) > 2 {
		switch args[2] {
		case "upsert_client":
			log.WithContext(ctx).Info("message subscribe upsert client started")
			done := make(chan struct{})
			go func() {
				err := rabbitmq.Subscriber(
					model.UpsertClientMessage,
					rabbitmq.KindFanOut,
					os.Getenv("UPSERT_CLIENT_MESSAGE_SUBSCRIBE"),
					"",
					func(msg []byte) bool {
						return port.Client().Upsert(msg)
					},
				)
				if err != nil {
					log.WithContext(ctx).Errorf("failed to subscribe to %s: %s", model.UpsertClientMessage, err)
				}
				close(done)
			}()
			<-done
			
		case "pppoe_create":
			log.WithContext(ctx).Info("PPPoE create consumer started")
			consumer := NewPppoeSyncConsumer(domainRegistry)
			done := make(chan struct{})
			go func() {
				err := rabbitmq.Subscriber(
					"mikrotik.ppp.create",
					rabbitmq.KindDirect,
					"mikrotik_ppp_create",
					"",
					consumer.HandleCreate,
				)
				if err != nil {
					log.WithContext(ctx).Errorf("failed to subscribe to mikrotik.ppp.create: %s", err)
				}
				close(done)
			}()
			<-done
			
		case "pppoe_isolate":
			log.WithContext(ctx).Info("PPPoE isolate consumer started")
			consumer := NewPppoeSyncConsumer(domainRegistry)
			done := make(chan struct{})
			go func() {
				err := rabbitmq.Subscriber(
					"mikrotik.ppp.isolate",
					rabbitmq.KindDirect,
					"mikrotik_ppp_isolate",
					"",
					consumer.HandleIsolate,
				)
				if err != nil {
					log.WithContext(ctx).Errorf("failed to subscribe to mikrotik.ppp.isolate: %s", err)
				}
				close(done)
			}()
			<-done
			
		case "pppoe_restore":
			log.WithContext(ctx).Info("PPPoE restore consumer started")
			consumer := NewPppoeSyncConsumer(domainRegistry)
			done := make(chan struct{})
			go func() {
				err := rabbitmq.Subscriber(
					"mikrotik.ppp.restore",
					rabbitmq.KindDirect,
					"mikrotik_ppp_restore",
					"",
					consumer.HandleRestore,
				)
				if err != nil {
					log.WithContext(ctx).Errorf("failed to subscribe to mikrotik.ppp.restore: %s", err)
				}
				close(done)
			}()
			<-done
			
		case "invoice_paid":
			log.WithContext(ctx).Info("Invoice paid consumer started")
			consumer := NewInvoicePaidConsumer(domainRegistry)
			done := make(chan struct{})
			go func() {
				err := rabbitmq.Subscriber(
					"invoice.paid",
					rabbitmq.KindDirect,
					"invoice_paid",
					"",
					consumer.Handle,
				)
				if err != nil {
					log.WithContext(ctx).Errorf("failed to subscribe to invoice.paid: %s", err)
				}
				close(done)
			}()
			<-done
			
		default:
			log.WithContext(ctx).Info("message subscribe not found")
		}
	} else {
		log.WithContext(ctx).Info("message subscribe not found")
	}
}
