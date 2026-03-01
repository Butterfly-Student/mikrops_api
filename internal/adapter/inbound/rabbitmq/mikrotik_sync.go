package rabbitmq_inbound_adapter

import (
	"context"
	"encoding/json"
	"fmt"

	"go-template/internal/model"
	inbound_port "go-template/internal/port/inbound"
	"go-template/pkg/hotspot"
	outbound_port "go-template/internal/port/outbound"
	"go-template/utils/activity"
	"go-template/utils/log"
)

type mikrotikSyncAdapter struct {
	mikrotikPort outbound_port.MikrotikPort
	hotspotPort  outbound_port.HotspotPort
	dbPort       outbound_port.DatabasePort
}

// NewMikrotikSyncAdapter creates a new RabbitMQ consumer adapter for MikroTik sync
func NewMikrotikSyncAdapter(mikrotikPort outbound_port.MikrotikPort, hotspotPort outbound_port.HotspotPort, dbPort outbound_port.DatabasePort) inbound_port.MikrotikSyncMessagePort {
	return &mikrotikSyncAdapter{
		mikrotikPort: mikrotikPort,
		hotspotPort:  hotspotPort,
		dbPort:       dbPort,
	}
}

// ProcessPPPoESync processes PPPoE sync messages
func (h *mikrotikSyncAdapter) ProcessPPPoESync(data []byte) bool {
	ctx := activity.NewContext("mikrotik_sync_pppoe")
	
	var message model.MikrotikSyncMessage
	if err := json.Unmarshal(data, &message); err != nil {
		log.WithContext(ctx).Error("Failed to unmarshal PPPoE sync message", err)
		return true // Ack message to prevent requeue of malformed message
	}
	
	ctx = context.WithValue(ctx, activity.Payload, message)
	
	// Get router details
	router, err := h.dbPort.Mikrotik().FindByID(message.RouterID.String())
	if err != nil {
		log.WithContext(ctx).Error("Failed to get router for PPPoE sync", err)
		return false // Nack - will retry
	}
	
	// Process based on action
	switch message.Action {
	case model.MikrotikSyncActionCreate:
		err = h.createPPPoESecret(ctx, router, message.PPPoEData)
	case model.MikrotikSyncActionUpdate:
		err = h.updatePPPoESecret(ctx, router, message.PPPoEData)
	case model.MikrotikSyncActionDelete:
		err = h.deletePPPoESecret(ctx, router, message.PPPoEData)
	case model.MikrotikSyncActionDisable:
		err = h.disablePPPoESecret(ctx, router, message.PPPoEData)
	case model.MikrotikSyncActionEnable:
		err = h.enablePPPoESecret(ctx, router, message.PPPoEData)
	default:
		log.WithContext(ctx).Warn("Unknown PPPoE sync action: " + string(message.Action))
		return true
	}
	
	if err != nil {
		log.WithContext(ctx).Error("PPPoE sync operation failed", err)
		return false // Nack - will retry
	}
	
	// Update subscription sync status
	if err := h.updateSubscriptionSyncStatus(ctx, message.SubscriptionID, true, ""); err != nil {
		log.WithContext(ctx).Error("Failed to update subscription sync status", err)
		// Don't return false here, the sync itself succeeded
	}
	
	log.WithContext(ctx).Info("PPPoE sync completed successfully")
	return true
}

// ProcessHotspotSync processes Hotspot sync messages
func (h *mikrotikSyncAdapter) ProcessHotspotSync(data []byte) bool {
	ctx := activity.NewContext("mikrotik_sync_hotspot")
	
	var message model.MikrotikSyncMessage
	if err := json.Unmarshal(data, &message); err != nil {
		log.WithContext(ctx).Error("Failed to unmarshal Hotspot sync message", err)
		return true
	}
	
	ctx = context.WithValue(ctx, activity.Payload, message)
	
	// Get router details
	router, err := h.dbPort.Mikrotik().FindByID(message.RouterID.String())
	if err != nil {
		log.WithContext(ctx).Error("Failed to get router for Hotspot sync", err)
		return false
	}
	
	// Process based on action
	switch message.Action {
	case model.MikrotikSyncActionCreate:
		err = h.createHotspotUser(ctx, router, message.HotspotData)
	case model.MikrotikSyncActionUpdate:
		err = h.updateHotspotUser(ctx, router, message.HotspotData)
	case model.MikrotikSyncActionDelete:
		err = h.deleteHotspotUser(ctx, router, message.HotspotData)
	case model.MikrotikSyncActionDisable:
		err = h.disableHotspotUser(ctx, router, message.HotspotData)
	case model.MikrotikSyncActionEnable:
		err = h.enableHotspotUser(ctx, router, message.HotspotData)
	default:
		log.WithContext(ctx).Warn("Unknown Hotspot sync action: " + string(message.Action))
		return true
	}
	
	if err != nil {
		log.WithContext(ctx).Error("Hotspot sync operation failed", err)
		return false
	}
	
	// Update subscription sync status
	if err := h.updateSubscriptionSyncStatus(ctx, message.SubscriptionID, true, ""); err != nil {
		log.WithContext(ctx).Error("Failed to update subscription sync status", err)
	}
	
	log.WithContext(ctx).Info("Hotspot sync completed successfully")
	return true
}

// Helper methods for PPPoE operations
func (h *mikrotikSyncAdapter) createPPPoESecret(ctx context.Context, router *model.MikrotikRouter, data *model.PPPoESyncData) error {
	secret := &model.PppoeSecret{
		Name:          data.Username,
		Password:      data.Password,
		Service:       "pppoe",
		Profile:       data.Profile,
		LocalAddress:  data.LocalAddress,
		RemoteAddress: data.RemoteAddress,
		Comment:       data.Comment,
	}
	return h.mikrotikPort.CreateSecret(router, secret)
}

func (h *mikrotikSyncAdapter) updatePPPoESecret(ctx context.Context, router *model.MikrotikRouter, data *model.PPPoESyncData) error {
	secret := &model.PppoeSecret{
		Name:          data.Username,
		Password:      data.Password,
		Service:       "pppoe",
		Profile:       data.Profile,
		LocalAddress:  data.LocalAddress,
		RemoteAddress: data.RemoteAddress,
		Comment:       data.Comment,
	}
	return h.mikrotikPort.UpdateSecret(router, secret)
}

func (h *mikrotikSyncAdapter) deletePPPoESecret(ctx context.Context, router *model.MikrotikRouter, data *model.PPPoESyncData) error {
	return h.mikrotikPort.DeleteSecret(router, data.Username)
}

func (h *mikrotikSyncAdapter) disablePPPoESecret(ctx context.Context, router *model.MikrotikRouter, data *model.PPPoESyncData) error {
	// Get existing secret first
	secret, err := h.mikrotikPort.GetSecret(router, data.Username)
	if err != nil {
		return err
	}
	secret.Disabled = true
	return h.mikrotikPort.UpdateSecret(router, secret)
}

func (h *mikrotikSyncAdapter) enablePPPoESecret(ctx context.Context, router *model.MikrotikRouter, data *model.PPPoESyncData) error {
	secret, err := h.mikrotikPort.GetSecret(router, data.Username)
	if err != nil {
		return err
	}
	secret.Disabled = false
	return h.mikrotikPort.UpdateSecret(router, secret)
}

// Helper methods for Hotspot operations
func (h *mikrotikSyncAdapter) createHotspotUser(ctx context.Context, router *model.MikrotikRouter, data *model.HotspotSyncData) error {
	hotspotClient, err := h.hotspotPort.GetHotspotClient(router)
	if err != nil {
		return err
	}
	
	user := &hotspot.User{
		Name:     data.Username,
		Password: data.Password,
		Profile:  data.Profile,
		Comment:  data.Comment,
		Server:   "all",
	}
	return hotspotClient.CreateUser(ctx, user)
}

func (h *mikrotikSyncAdapter) updateHotspotUser(ctx context.Context, router *model.MikrotikRouter, data *model.HotspotSyncData) error {
	hotspotClient, err := h.hotspotPort.GetHotspotClient(router)
	if err != nil {
		return err
	}
	
	updates := &hotspot.UserUpdate{
		Profile:  &data.Profile,
		Disabled: boolPtr(false),
	}
	return hotspotClient.UpdateUser(ctx, data.Username, updates)
}

func (h *mikrotikSyncAdapter) deleteHotspotUser(ctx context.Context, router *model.MikrotikRouter, data *model.HotspotSyncData) error {
	hotspotClient, err := h.hotspotPort.GetHotspotClient(router)
	if err != nil {
		return err
	}
	return hotspotClient.DeleteUser(ctx, data.Username)
}

func (h *mikrotikSyncAdapter) disableHotspotUser(ctx context.Context, router *model.MikrotikRouter, data *model.HotspotSyncData) error {
	hotspotClient, err := h.hotspotPort.GetHotspotClient(router)
	if err != nil {
		return err
	}
	return hotspotClient.DisableUser(ctx, data.Username)
}

func (h *mikrotikSyncAdapter) enableHotspotUser(ctx context.Context, router *model.MikrotikRouter, data *model.HotspotSyncData) error {
	hotspotClient, err := h.hotspotPort.GetHotspotClient(router)
	if err != nil {
		return err
	}
	return hotspotClient.EnableUser(ctx, data.Username)
}

// updateSubscriptionSyncStatus updates the sync status in the database
func (h *mikrotikSyncAdapter) updateSubscriptionSyncStatus(ctx context.Context, subscriptionID interface{}, synced bool, errorMsg string) error {
	// This would typically call a subscription domain method
	// For now, we just log it
	log.WithContext(ctx).Info("Updating subscription sync status: subscription_id=" + subscriptionID.(string) + " synced=" + fmt.Sprintf("%v", synced))
	return nil
}

// Helper function
func boolPtr(b bool) *bool {
	return &b
}
