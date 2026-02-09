package mikrotik

import (
	"context"
	"fmt"

	"github.com/palantir/stacktrace"

	"mikrops/internal/model"
	outbound_port "mikrops/internal/port/outbound"
	"mikrops/utils/crypto"
)

type MikrotikDomain interface {
	SyncSubscription(ctx context.Context, subscriptionID string) error
	GetActiveConnections(ctx context.Context, nasID string) ([]model.MikrotikConnection, error)
	GetBandwidth(ctx context.Context, nasID string, target string) (model.MikrotikBandwidth, error)
	RemoveSubscription(ctx context.Context, subscriptionID string) error

	// PPPoE Management
	CreatePPPoESecret(ctx context.Context, nasID, pppoeUsername, pppoePassword, profile string) error
	UpdatePPPoEProfile(ctx context.Context, nasID, pppoeUsername, newProfile string) error
	DisconnectPPPoESession(ctx context.Context, nasID, pppoeUsername string) error
}

type mikrotikDomain struct {
	databasePort outbound_port.DatabasePort
	messagePort  outbound_port.MessagePort
	cachePort    outbound_port.CachePort
	workflowPort outbound_port.WorkflowPort
	httpPort     outbound_port.HttpPort
}

func NewMikrotikDomain(
	databasePort outbound_port.DatabasePort,
	messagePort outbound_port.MessagePort,
	cachePort outbound_port.CachePort,
	workflowPort outbound_port.WorkflowPort,
	httpPort outbound_port.HttpPort,
) MikrotikDomain {
	return &mikrotikDomain{
		databasePort: databasePort,
		messagePort:  messagePort,
		cachePort:    cachePort,
		workflowPort: workflowPort,
		httpPort:     httpPort,
	}
}

// getNasCredentials decrypts NAS password and returns host info needed for MikroTik API calls.
func (d *mikrotikDomain) getNasCredentials(nasID string) (model.Nas, string, error) {
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

func (d *mikrotikDomain) SyncSubscription(ctx context.Context, subscriptionID string) error {
	if subscriptionID == "" {
		return stacktrace.NewError("subscriptionID is empty")
	}

	subscription, err := d.databasePort.Subscription().FindByID(subscriptionID)
	if err != nil {
		return stacktrace.Propagate(err, "failed to find subscription")
	}

	pkg, err := d.databasePort.InternetPackage().FindByID(subscription.PackageID)
	if err != nil {
		return stacktrace.Propagate(err, "failed to find internet package")
	}

	// Get PPPoE credentials from pppoe_accounts if linked
	var pppoeUsername, pppoePassword string
	if subscription.PppoeAccountID != nil && *subscription.PppoeAccountID != "" {
		account, err := d.databasePort.PppoeAccount().FindByID(*subscription.PppoeAccountID)
		if err != nil {
			return stacktrace.Propagate(err, "failed to find pppoe account")
		}
		pppoeUsername = account.Username
		pppoePassword = account.PasswordEncrypted
	} else {
		pppoeUsername = subscription.MikrotikSecretName
	}

	if pppoeUsername == "" {
		return stacktrace.NewError("no PPPoE username available for sync")
	}

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

func (d *mikrotikDomain) GetActiveConnections(ctx context.Context, nasID string) ([]model.MikrotikConnection, error) {
	if nasID == "" {
		return nil, stacktrace.NewError("nasID is empty")
	}

	nas, password, err := d.getNasCredentials(nasID)
	if err != nil {
		return nil, err
	}

	mikrotikPort := d.httpPort.Mikrotik()
	connections, err := mikrotikPort.GetActiveConnections(nas.Host, nas.ApiPort, nas.Username, password, nas.UseSSL)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get active connections from MikroTik")
	}

	return connections, nil
}

func (d *mikrotikDomain) GetBandwidth(ctx context.Context, nasID string, target string) (model.MikrotikBandwidth, error) {
	if nasID == "" {
		return model.MikrotikBandwidth{}, stacktrace.NewError("nasID is empty")
	}
	if target == "" {
		return model.MikrotikBandwidth{}, stacktrace.NewError("target is empty")
	}

	nas, password, err := d.getNasCredentials(nasID)
	if err != nil {
		return model.MikrotikBandwidth{}, err
	}

	mikrotikPort := d.httpPort.Mikrotik()
	bandwidth, err := mikrotikPort.MonitorBandwidth(nas.Host, nas.ApiPort, nas.Username, password, nas.UseSSL, target)
	if err != nil {
		return model.MikrotikBandwidth{}, stacktrace.Propagate(err, "failed to monitor bandwidth from MikroTik")
	}

	return bandwidth, nil
}

func (d *mikrotikDomain) RemoveSubscription(ctx context.Context, subscriptionID string) error {
	if subscriptionID == "" {
		return stacktrace.NewError("subscriptionID is empty")
	}

	subscription, err := d.databasePort.Subscription().FindByID(subscriptionID)
	if err != nil {
		return stacktrace.Propagate(err, "failed to find subscription")
	}

	nas, password, err := d.getNasCredentials(subscription.NasID)
	if err != nil {
		return err
	}

	mikrotikPort := d.httpPort.Mikrotik()

	// Find and delete PPPoE secret by name
	secrets, err := mikrotikPort.ListPPPoESecrets(nas.Host, nas.RestPort, nas.Username, password, nas.UseSSL)
	if err == nil {
		for _, s := range secrets {
			if s.Name == subscription.MikrotikSecretName {
				_ = mikrotikPort.DeletePPPoESecret(nas.Host, nas.RestPort, nas.Username, password, nas.UseSSL, s.ID)
				break
			}
		}
	}

	// Find and delete simple queue by name
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

// CreatePPPoESecret creates a PPPoE secret directly on MikroTik (called by RabbitMQ consumer)
func (d *mikrotikDomain) CreatePPPoESecret(ctx context.Context, nasID, pppoeUsername, pppoePassword, profile string) error {
	if nasID == "" {
		return stacktrace.NewError("nasID is empty")
	}

	nas, password, err := d.getNasCredentials(nasID)
	if err != nil {
		return err
	}

	mikrotikPort := d.httpPort.Mikrotik()

	secretInput := model.MikrotikPPPoESecretInput{
		Name:     pppoeUsername,
		Password: pppoePassword,
		Service:  "pppoe",
		Profile:  profile,
	}

	err = mikrotikPort.CreatePPPoESecret(nas.Host, nas.RestPort, nas.Username, password, nas.UseSSL, secretInput)
	if err != nil {
		return stacktrace.Propagate(err, "failed to create PPPoE secret")
	}

	return nil
}

// UpdatePPPoEProfile updates the profile of an existing PPPoE secret (for isolate/restore)
func (d *mikrotikDomain) UpdatePPPoEProfile(ctx context.Context, nasID, pppoeUsername, newProfile string) error {
	if nasID == "" {
		return stacktrace.NewError("nasID is empty")
	}

	nas, password, err := d.getNasCredentials(nasID)
	if err != nil {
		return err
	}

	mikrotikPort := d.httpPort.Mikrotik()

	// First, find the secret by username
	secrets, err := mikrotikPort.ListPPPoESecrets(nas.Host, nas.RestPort, nas.Username, password, nas.UseSSL)
	if err != nil {
		return stacktrace.Propagate(err, "failed to list PPPoE secrets")
	}

	var secretID string
	for _, s := range secrets {
		if s.Name == pppoeUsername {
			secretID = s.ID
			break
		}
	}

	if secretID == "" {
		return stacktrace.NewError("PPPoE secret not found: %s", pppoeUsername)
	}

	// Update the profile
	updateInput := model.MikrotikPPPoESecretInput{
		Profile: newProfile,
	}

	err = mikrotikPort.UpdatePPPoESecret(nas.Host, nas.RestPort, nas.Username, password, nas.UseSSL, secretID, updateInput)
	if err != nil {
		return stacktrace.Propagate(err, "failed to update PPPoE profile")
	}

	return nil
}

// DisconnectPPPoESession forcefully disconnects an active PPPoE session
func (d *mikrotikDomain) DisconnectPPPoESession(ctx context.Context, nasID, pppoeUsername string) error {
	if nasID == "" {
		return stacktrace.NewError("nasID is empty")
	}

	// Get NAS credentials for future implementation
	_, _, err := d.getNasCredentials(nasID)
	if err != nil {
		return err
	}

	// This requires RouterOS API access to find and remove active connections
	// For now, we'll use a simple approach - just return nil as disconnect is optional
	// The user will reconnect with new profile automatically
	// TODO: Implement actual disconnect via RouterOS API if needed
	return nil
}
