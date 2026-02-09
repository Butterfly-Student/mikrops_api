package subscription

import (
	"context"
	"fmt"
	"time"

	"github.com/palantir/stacktrace"

	"mikrops/internal/model"
	outbound_port "mikrops/internal/port/outbound"
	"mikrops/utils/crypto"
)

type SubscriptionDomain interface {
	Create(ctx context.Context, input model.SubscriptionInput) (model.Subscription, error)
	FindByFilter(ctx context.Context, filter model.SubscriptionFilter) ([]model.Subscription, error)
	FindByID(ctx context.Context, id string) (model.Subscription, error)
	Update(ctx context.Context, id string, input model.SubscriptionInput) error
	Activate(ctx context.Context, id string) error
	Suspend(ctx context.Context, id string) error
	SetVacation(ctx context.Context, id string, start, end time.Time) error
	Cancel(ctx context.Context, id string) error
	FindExpiring(ctx context.Context, days int, tenantID string) ([]model.Subscription, error)
}

type subscriptionDomain struct {
	databasePort outbound_port.DatabasePort
	messagePort  outbound_port.MessagePort
	cachePort    outbound_port.CachePort
	workflowPort outbound_port.WorkflowPort
	httpPort     outbound_port.HttpPort
}

func NewSubscriptionDomain(
	databasePort outbound_port.DatabasePort,
	messagePort outbound_port.MessagePort,
	cachePort outbound_port.CachePort,
	workflowPort outbound_port.WorkflowPort,
	httpPort outbound_port.HttpPort,
) SubscriptionDomain {
	return &subscriptionDomain{
		databasePort: databasePort,
		messagePort:  messagePort,
		cachePort:    cachePort,
		workflowPort: workflowPort,
		httpPort:     httpPort,
	}
}

func (d *subscriptionDomain) getNasCredentials(nasID string) (model.Nas, string, error) {
	nas, err := d.databasePort.Nas().FindByID(nasID)
	if err != nil {
		return model.Nas{}, "", stacktrace.Propagate(err, "failed to find NAS")
	}

	password, err := crypto.Decrypt(nas.PasswordEncrypted)
	if err != nil {
		return model.Nas{}, "", stacktrace.Propagate(err, "failed to decrypt NAS password")
	}

	return nas, password, nil
}

func (d *subscriptionDomain) getPppoeUsername(subscription model.Subscription) string {
	// Try to get PPPoE username from linked PPPoE account
	if subscription.PppoeAccountID != nil && *subscription.PppoeAccountID != "" {
		account, err := d.databasePort.PppoeAccount().FindByID(*subscription.PppoeAccountID)
		if err == nil {
			return account.Username
		}
	}
	// Fallback to MikrotikSecretName
	return subscription.MikrotikSecretName
}

func (d *subscriptionDomain) syncToMikrotik(subscription model.Subscription, pppoeUsername string, pppoePassword string, pkg model.InternetPackage) error {
	nas, password, err := d.getNasCredentials(subscription.NasID)
	if err != nil {
		return err
	}

	mikrotikPort := d.httpPort.Mikrotik()

	// Determine profile name
	profileName := pkg.Name
	if pkg.ProfileName != "" {
		profileName = pkg.ProfileName
	}

	// Create/update PPPoE secret
	secretInput := model.MikrotikPPPoESecretInput{
		Name:     pppoeUsername,
		Password: pppoePassword,
		Service:  "pppoe",
		Profile:  profileName,
		Comment:  fmt.Sprintf("sub:%s", subscription.ID),
	}
	err = mikrotikPort.CreatePPPoESecret(nas.Host, nas.RestPort, nas.Username, password, nas.UseSSL, secretInput)
	if err != nil {
		return stacktrace.Propagate(err, "failed to create PPPoE secret in MikroTik")
	}

	// Create simple queue for bandwidth control
	maxLimit := fmt.Sprintf("%s/%s", pkg.UploadRate, pkg.DownloadRate)
	burstLimit := ""
	if pkg.UploadBurst != "" && pkg.DownloadBurst != "" {
		burstLimit = fmt.Sprintf("%s/%s", pkg.UploadBurst, pkg.DownloadBurst)
	}
	queueInput := model.MikrotikSimpleQueueInput{
		Name:       subscription.MikrotikQueueName,
		Target:     fmt.Sprintf("<pppoe-%s>", pppoeUsername),
		MaxLimit:   maxLimit,
		BurstLimit: burstLimit,
		Comment:    fmt.Sprintf("sub:%s", subscription.ID),
	}
	err = mikrotikPort.CreateSimpleQueue(nas.Host, nas.RestPort, nas.Username, password, nas.UseSSL, queueInput)
	if err != nil {
		return stacktrace.Propagate(err, "failed to create simple queue in MikroTik")
	}

	return nil
}

func (d *subscriptionDomain) findAndTogglePPPoESecret(subscription model.Subscription, disable bool) error {
	nas, password, err := d.getNasCredentials(subscription.NasID)
	if err != nil {
		return err
	}

	mikrotikPort := d.httpPort.Mikrotik()
	secrets, err := mikrotikPort.ListPPPoESecrets(nas.Host, nas.RestPort, nas.Username, password, nas.UseSSL)
	if err != nil {
		return stacktrace.Propagate(err, "failed to list PPPoE secrets")
	}

	for _, s := range secrets {
		if s.Name == subscription.MikrotikSecretName {
			if disable {
				return mikrotikPort.DisablePPPoESecret(nas.Host, nas.RestPort, nas.Username, password, nas.UseSSL, s.ID)
			}
			return mikrotikPort.EnablePPPoESecret(nas.Host, nas.RestPort, nas.Username, password, nas.UseSSL, s.ID)
		}
	}

	return nil
}

func (d *subscriptionDomain) removeFromMikrotik(subscription model.Subscription) error {
	nas, password, err := d.getNasCredentials(subscription.NasID)
	if err != nil {
		return err
	}

	mikrotikPort := d.httpPort.Mikrotik()

	// Delete PPPoE secret
	secrets, err := mikrotikPort.ListPPPoESecrets(nas.Host, nas.RestPort, nas.Username, password, nas.UseSSL)
	if err == nil {
		for _, s := range secrets {
			if s.Name == subscription.MikrotikSecretName {
				_ = mikrotikPort.DeletePPPoESecret(nas.Host, nas.RestPort, nas.Username, password, nas.UseSSL, s.ID)
				break
			}
		}
	}

	// Delete simple queue
	queues, err := mikrotikPort.ListSimpleQueues(nas.Host, nas.RestPort, nas.Username, password, nas.UseSSL)
	if err == nil {
		for _, q := range queues {
			if q.Name == subscription.MikrotikQueueName {
				_ = mikrotikPort.DeleteSimpleQueue(nas.Host, nas.RestPort, nas.Username, password, nas.UseSSL, q.ID)
				break
			}
		}
	}

	return nil
}

func (d *subscriptionDomain) Create(ctx context.Context, input model.SubscriptionInput) (model.Subscription, error) {
	// Get internet package for queue configuration
	pkg, err := d.databasePort.InternetPackage().FindByID(input.PackageID)
	if err != nil {
		return model.Subscription{}, stacktrace.Propagate(err, "failed to find internet package")
	}

	// Set default status
	if input.Status == "" {
		input.Status = model.SubscriptionStatusActive
	}

	// Get PPPoE credentials from pppoe_accounts if linked
	var pppoeUsername, pppoePassword string
	if input.PppoeAccountID != nil && *input.PppoeAccountID != "" {
		account, err := d.databasePort.PppoeAccount().FindByID(*input.PppoeAccountID)
		if err != nil {
			return model.Subscription{}, stacktrace.Propagate(err, "failed to find pppoe account")
		}
		pppoeUsername = account.Username
		pppoePassword = account.PasswordEncrypted
		input.MikrotikSecretName = account.Username
		input.MikrotikQueueName = fmt.Sprintf("queue_%s", account.Username)
	} else {
		// Fallback: use subscription ID-based names
		input.MikrotikSecretName = fmt.Sprintf("sub_%s", input.CustomerID[:8])
		input.MikrotikQueueName = fmt.Sprintf("queue_sub_%s", input.CustomerID[:8])
	}

	// Create subscription
	subscription, err := d.databasePort.Subscription().Create(input)
	if err != nil {
		return model.Subscription{}, stacktrace.Propagate(err, "failed to create subscription")
	}

	// Sync to MikroTik if we have PPPoE credentials
	if pppoeUsername != "" {
		if err := d.syncToMikrotik(subscription, pppoeUsername, pppoePassword, pkg); err != nil {
			return subscription, stacktrace.Propagate(err, "failed to sync subscription to MikroTik")
		}
	}

	return subscription, nil
}

func (d *subscriptionDomain) FindByFilter(ctx context.Context, filter model.SubscriptionFilter) ([]model.Subscription, error) {
	if filter.IsEmpty() {
		return nil, stacktrace.NewError("filter is empty")
	}

	subscriptions, err := d.databasePort.Subscription().FindByFilter(filter)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to find subscriptions by filter")
	}

	return subscriptions, nil
}

func (d *subscriptionDomain) FindByID(ctx context.Context, id string) (model.Subscription, error) {
	if id == "" {
		return model.Subscription{}, stacktrace.NewError("id is empty")
	}

	subscription, err := d.databasePort.Subscription().FindByID(id)
	if err != nil {
		return model.Subscription{}, stacktrace.Propagate(err, "failed to find subscription by id")
	}

	return subscription, nil
}

func (d *subscriptionDomain) Update(ctx context.Context, id string, input model.SubscriptionInput) error {
	if id == "" {
		return stacktrace.NewError("id is empty")
	}

	err := d.databasePort.Subscription().Update(id, input)
	if err != nil {
		return stacktrace.Propagate(err, "failed to update subscription")
	}

	return nil
}

func (d *subscriptionDomain) Activate(ctx context.Context, id string) error {
	if id == "" {
		return stacktrace.NewError("id is empty")
	}

	subscription, err := d.databasePort.Subscription().FindByID(id)
	if err != nil {
		return stacktrace.Propagate(err, "failed to find subscription")
	}

	subscription.Status = model.SubscriptionStatusActive
	err = d.databasePort.Subscription().Update(id, subscription.SubscriptionInput)
	if err != nil {
		return stacktrace.Propagate(err, "failed to activate subscription")
	}

	// Enable PPPoE secret in MikroTik
	if err := d.findAndTogglePPPoESecret(subscription, false); err != nil {
		return stacktrace.Propagate(err, "failed to enable PPPoE secret in MikroTik")
	}

	return nil
}

func (d *subscriptionDomain) Suspend(ctx context.Context, id string) error {
	if id == "" {
		return stacktrace.NewError("id is empty")
	}

	subscription, err := d.databasePort.Subscription().FindByID(id)
	if err != nil {
		return stacktrace.Propagate(err, "failed to find subscription")
	}

	subscription.Status = model.SubscriptionStatusSuspended
	err = d.databasePort.Subscription().Update(id, subscription.SubscriptionInput)
	if err != nil {
		return stacktrace.Propagate(err, "failed to suspend subscription")
	}

	// Disable PPPoE secret in MikroTik
	if err := d.findAndTogglePPPoESecret(subscription, true); err != nil {
		return stacktrace.Propagate(err, "failed to disable PPPoE secret in MikroTik")
	}

	return nil
}

func (d *subscriptionDomain) SetVacation(ctx context.Context, id string, start, end time.Time) error {
	if id == "" {
		return stacktrace.NewError("id is empty")
	}

	subscription, err := d.databasePort.Subscription().FindByID(id)
	if err != nil {
		return stacktrace.Propagate(err, "failed to find subscription")
	}

	subscription.Status = model.SubscriptionStatusVacation
	subscription.VacationStart = &start
	subscription.VacationEnd = &end
	err = d.databasePort.Subscription().Update(id, subscription.SubscriptionInput)
	if err != nil {
		return stacktrace.Propagate(err, "failed to set vacation")
	}

	// Disable PPPoE secret in MikroTik during vacation
	if err := d.findAndTogglePPPoESecret(subscription, true); err != nil {
		return stacktrace.Propagate(err, "failed to disable PPPoE secret in MikroTik for vacation")
	}

	return nil
}

func (d *subscriptionDomain) Cancel(ctx context.Context, id string) error {
	if id == "" {
		return stacktrace.NewError("id is empty")
	}

	subscription, err := d.databasePort.Subscription().FindByID(id)
	if err != nil {
		return stacktrace.Propagate(err, "failed to find subscription")
	}

	subscription.Status = model.SubscriptionStatusCancelled
	err = d.databasePort.Subscription().Update(id, subscription.SubscriptionInput)
	if err != nil {
		return stacktrace.Propagate(err, "failed to cancel subscription")
	}

	// Remove PPPoE secret and queue from MikroTik
	if err := d.removeFromMikrotik(subscription); err != nil {
		return stacktrace.Propagate(err, "failed to remove subscription from MikroTik")
	}

	return nil
}

func (d *subscriptionDomain) FindExpiring(ctx context.Context, days int, tenantID string) ([]model.Subscription, error) {
	if days <= 0 {
		return nil, stacktrace.NewError("days must be positive")
	}

	subscriptions, err := d.databasePort.Subscription().FindExpiring(days, tenantID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to find expiring subscriptions")
	}

	return subscriptions, nil
}
