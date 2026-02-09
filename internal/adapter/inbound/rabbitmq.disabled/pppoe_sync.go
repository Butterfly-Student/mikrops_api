package rabbitmq_inbound_adapter

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/palantir/stacktrace"

	"mikrops/internal/domain"
	"mikrops/internal/model"
	"mikrops/utils/log"
)

// PppoeSyncConsumer handles PPPoE account synchronization messages
type PppoeSyncConsumer struct {
	domain domain.Domain
}

func NewPppoeSyncConsumer(d domain.Domain) *PppoeSyncConsumer {
	return &PppoeSyncConsumer{domain: d}
}

// HandleCreate processes mikrotik.ppp.create messages
func (c *PppoeSyncConsumer) HandleCreate(msg []byte) bool {
	ctx := context.Background()
	
	var message struct {
		Action         string `json:"action"`
		PppoeAccountID string `json:"pppoe_account_id"`
		NasID          string `json:"nas_id"`
		Username       string `json:"username"`
		Password       string `json:"password"`
		Profile        string `json:"profile"`
	}
	
	err := json.Unmarshal(msg, &message)
	if err != nil {
		log.WithContext(ctx).Errorf("failed to unmarshal create message: %v", err)
		return false
	}
	
	// Get PPPoE account
	account, err := c.domain.PppoeAccount().FindByID(ctx, message.PppoeAccountID)
	if err != nil {
		log.WithContext(ctx).Errorf("failed to find pppoe account %s: %v", message.PppoeAccountID, err)
		c.logSyncError(ctx, message.PppoeAccountID, message.NasID, "create", err.Error())
		return false
	}
	
	// Get NAS details
	nas, err := c.domain.Nas().FindByID(ctx, message.NasID)
	if err != nil {
		log.WithContext(ctx).Errorf("failed to find nas %s: %v", message.NasID, err)
		c.logSyncError(ctx, message.PppoeAccountID, message.NasID, "create", err.Error())
		return false
	}
	
	// Call MikroTik to create PPPoE secret
	err = c.domain.Mikrotik().CreatePPPoESecret(ctx, nas.Host, nas.Username, nas.Password, 
		message.Username, message.Password, message.Profile)
	if err != nil {
		log.WithContext(ctx).Errorf("failed to create pppoe secret on mikrotik: %v", err)
		c.logSyncError(ctx, message.PppoeAccountID, message.NasID, "create", err.Error())
		
		// Update sync status to error
		_ = c.domain.PppoeAccount().Update(ctx, message.PppoeAccountID, model.PppoeAccountInput{
			SyncStatus: model.PppoeSyncStatusError,
		})
		return false
	}
	
	// Update sync status to synced
	_ = c.domain.PppoeAccount().Update(ctx, message.PppoeAccountID, model.PppoeAccountInput{
		SyncStatus: model.PppoeSyncStatusSynced,
	})
	
	// Log success
	c.logSyncSuccess(ctx, message.PppoeAccountID, message.NasID, "create", 
		fmt.Sprintf("PPPoE secret created: %s", message.Username))
	
	log.WithContext(ctx).Infof("PPPoE secret created successfully: %s on NAS %s", message.Username, nas.Name)
	return true
}

// HandleIsolate processes mikrotik.ppp.isolate messages
func (c *PppoeSyncConsumer) HandleIsolate(msg []byte) bool {
	ctx := context.Background()
	
	var message struct {
		Action         string `json:"action"`
		PppoeAccountID string `json:"pppoe_account_id"`
		NasID          string `json:"nas_id"`
		Username       string `json:"username"`
		IsolirProfile  string `json:"isolir_profile"`
	}
	
	err := json.Unmarshal(msg, &message)
	if err != nil {
		log.WithContext(ctx).Errorf("failed to unmarshal isolate message: %v", err)
		return false
	}
	
	// Get PPPoE account
	account, err := c.domain.PppoeAccount().FindByID(ctx, message.PppoeAccountID)
	if err != nil {
		log.WithContext(ctx).Errorf("failed to find pppoe account %s: %v", message.PppoeAccountID, err)
		c.logSyncError(ctx, message.PppoeAccountID, message.NasID, "isolate", err.Error())
		return false
	}
	
	// Get NAS details
	nas, err := c.domain.Nas().FindByID(ctx, message.NasID)
	if err != nil {
		log.WithContext(ctx).Errorf("failed to find nas %s: %v", message.NasID, err)
		c.logSyncError(ctx, message.PppoeAccountID, message.NasID, "isolate", err.Error())
		return false
	}
	
	// Call MikroTik to change profile
	err = c.domain.Mikrotik().UpdatePPPoEProfile(ctx, nas.Host, nas.Username, nas.Password,
		message.Username, message.IsolirProfile)
	if err != nil {
		log.WithContext(ctx).Errorf("failed to update pppoe profile on mikrotik: %v", err)
		c.logSyncError(ctx, message.PppoeAccountID, message.NasID, "isolate", err.Error())
		
		// Update sync status to error
		_ = c.domain.PppoeAccount().Update(ctx, message.PppoeAccountID, model.PppoeAccountInput{
			SyncStatus: model.PppoeSyncStatusError,
		})
		return false
	}
	
	// Force disconnect to apply profile change
	_ = c.domain.Mikrotik().DisconnectPPPoESession(ctx, nas.Host, nas.Username, nas.Password, message.Username)
	
	// Update sync status to synced
	_ = c.domain.PppoeAccount().Update(ctx, message.PppoeAccountID, model.PppoeAccountInput{
		SyncStatus: model.PppoeSyncStatusSynced,
	})
	
	// Log success
	c.logSyncSuccess(ctx, message.PppoeAccountID, message.NasID, "isolate",
		fmt.Sprintf("PPPoE account isolated: %s -> %s", message.Username, message.IsolirProfile))
	
	log.WithContext(ctx).Infof("PPPoE account isolated successfully: %s on NAS %s", message.Username, nas.Name)
	return true
}

// HandleRestore processes mikrotik.ppp.restore messages
func (c *PppoeSyncConsumer) HandleRestore(msg []byte) bool {
	ctx := context.Background()
	
	var message struct {
		Action         string `json:"action"`
		PppoeAccountID string `json:"pppoe_account_id"`
		NasID          string `json:"nas_id"`
		Username       string `json:"username"`
		RestoreProfile string `json:"restore_profile"`
	}
	
	err := json.Unmarshal(msg, &message)
	if err != nil {
		log.WithContext(ctx).Errorf("failed to unmarshal restore message: %v", err)
		return false
	}
	
	// Get PPPoE account
	account, err := c.domain.PppoeAccount().FindByID(ctx, message.PppoeAccountID)
	if err != nil {
		log.WithContext(ctx).Errorf("failed to find pppoe account %s: %v", message.PppoeAccountID, err)
		c.logSyncError(ctx, message.PppoeAccountID, message.NasID, "restore", err.Error())
		return false
	}
	
	// Get NAS details
	nas, err := c.domain.Nas().FindByID(ctx, message.NasID)
	if err != nil {
		log.WithContext(ctx).Errorf("failed to find nas %s: %v", message.NasID, err)
		c.logSyncError(ctx, message.PppoeAccountID, message.NasID, "restore", err.Error())
		return false
	}
	
	// Call MikroTik to restore profile
	err = c.domain.Mikrotik().UpdatePPPoEProfile(ctx, nas.Host, nas.Username, nas.Password,
		message.Username, message.RestoreProfile)
	if err != nil {
		log.WithContext(ctx).Errorf("failed to restore pppoe profile on mikrotik: %v", err)
		c.logSyncError(ctx, message.PppoeAccountID, message.NasID, "restore", err.Error())
		
		// Update sync status to error
		_ = c.domain.PppoeAccount().Update(ctx, message.PppoeAccountID, model.PppoeAccountInput{
			SyncStatus: model.PppoeSyncStatusError,
		})
		return false
	}
	
	// Force disconnect to apply profile change
	_ = c.domain.Mikrotik().DisconnectPPPoESession(ctx, nas.Host, nas.Username, nas.Password, message.Username)
	
	// Update sync status to synced
	_ = c.domain.PppoeAccount().Update(ctx, message.PppoeAccountID, model.PppoeAccountInput{
		SyncStatus: model.PppoeSyncStatusSynced,
	})
	
	// Log success
	c.logSyncSuccess(ctx, message.PppoeAccountID, message.NasID, "restore",
		fmt.Sprintf("PPPoE account restored: %s -> %s", message.Username, message.RestoreProfile))
	
	log.WithContext(ctx).Infof("PPPoE account restored successfully: %s on NAS %s", message.Username, nas.Name)
	return true
}

// logSyncSuccess logs successful sync to database
func (c *PppoeSyncConsumer) logSyncSuccess(ctx context.Context, pppoeAccountID, nasID, action, response string) {
	account, err := c.domain.PppoeAccount().FindByID(ctx, pppoeAccountID)
	if err != nil {
		return
	}
	
	logInput := model.MikrotikSyncLogInput{
		TenantID:        account.TenantID,
		NasID:           nasID,
		PppoeAccountID:  &pppoeAccountID,
		Action:          action,
		Status:          "success",
		ResponsePayload: response,
	}
	_, _ = c.domain.MikrotikSyncLog().Create(ctx, logInput)
}

// logSyncError logs failed sync to database
func (c *PppoeSyncConsumer) logSyncError(ctx context.Context, pppoeAccountID, nasID, action, errorMsg string) {
	account, err := c.domain.PppoeAccount().FindByID(ctx, pppoeAccountID)
	if err != nil {
		// If we can't get account, create minimal log entry
		logInput := model.MikrotikSyncLogInput{
			NasID:          nasID,
			PppoeAccountID: &pppoeAccountID,
			Action:         action,
			Status:         "error",
			ErrorMessage:   errorMsg,
		}
		_, _ = c.domain.MikrotikSyncLog().Create(ctx, logInput)
		return
	}
	
	logInput := model.MikrotikSyncLogInput{
		TenantID:       account.TenantID,
		NasID:          nasID,
		PppoeAccountID: &pppoeAccountID,
		Action:         action,
		Status:         "error",
		ErrorMessage:   errorMsg,
	}
	_, _ = c.domain.MikrotikSyncLog().Create(ctx, logInput)
}
