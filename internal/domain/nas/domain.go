package nas

import (
	"context"

	"github.com/palantir/stacktrace"

	"mikrops/internal/model"
	outbound_port "mikrops/internal/port/outbound"
	"mikrops/utils/crypto"
)

type NasDomain interface {
	Create(ctx context.Context, input model.NasInput) (model.Nas, error)
	FindByFilter(ctx context.Context, filter model.NasFilter) ([]model.Nas, error)
	FindByID(ctx context.Context, id string) (model.Nas, error)
	Update(ctx context.Context, id string, input model.NasInput) error
	Delete(ctx context.Context, id string) error
	CountByTenantID(ctx context.Context, tenantID string) (int, error)
	TestConnection(ctx context.Context, id string) error
}

type nasDomain struct {
	databasePort outbound_port.DatabasePort
	messagePort  outbound_port.MessagePort
	cachePort    outbound_port.CachePort
	workflowPort outbound_port.WorkflowPort
	httpPort     outbound_port.HttpPort
}

func NewNasDomain(
	databasePort outbound_port.DatabasePort,
	messagePort outbound_port.MessagePort,
	cachePort outbound_port.CachePort,
	workflowPort outbound_port.WorkflowPort,
	httpPort outbound_port.HttpPort,
) NasDomain {
	return &nasDomain{
		databasePort: databasePort,
		messagePort:  messagePort,
		cachePort:    cachePort,
		workflowPort: workflowPort,
		httpPort:     httpPort,
	}
}

func (d *nasDomain) Create(ctx context.Context, input model.NasInput) (model.Nas, error) {
	// Check max NAS limit per tenant (max 3)
	count, err := d.databasePort.Nas().CountByTenantID(input.TenantID)
	if err != nil {
		return model.Nas{}, stacktrace.Propagate(err, "failed to count NAS for tenant")
	}

	// Get tenant to check max_nas limit
	tenant, err := d.databasePort.Tenant().FindByID(input.TenantID)
	if err != nil {
		return model.Nas{}, stacktrace.Propagate(err, "failed to find tenant")
	}

	maxNas := tenant.MaxNas
	if maxNas == 0 {
		maxNas = 3 // default
	}

	if count >= maxNas {
		return model.Nas{}, stacktrace.NewError("maximum number of NAS (%d) reached for this tenant", maxNas)
	}

	// Encrypt password
	if input.Password != "" {
		encryptedPassword, err := crypto.Encrypt(input.Password)
		if err != nil {
			return model.Nas{}, stacktrace.Propagate(err, "failed to encrypt password")
		}
		input.PasswordEncrypted = encryptedPassword
		input.Password = "" // Clear plain password
	}

	nas, err := d.databasePort.Nas().Create(input)
	if err != nil {
		return model.Nas{}, stacktrace.Propagate(err, "failed to create NAS")
	}

	// Provision isolir profile on MikroTik asynchronously
	go d.provisionIsolirProfile(context.Background(), nas)

	return nas, nil
}

func (d *nasDomain) FindByFilter(ctx context.Context, filter model.NasFilter) ([]model.Nas, error) {
	if filter.IsEmpty() {
		return nil, stacktrace.NewError("filter is empty")
	}

	nasList, err := d.databasePort.Nas().FindByFilter(filter)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to find NAS by filter")
	}

	return nasList, nil
}

func (d *nasDomain) FindByID(ctx context.Context, id string) (model.Nas, error) {
	if id == "" {
		return model.Nas{}, stacktrace.NewError("id is empty")
	}

	nas, err := d.databasePort.Nas().FindByID(id)
	if err != nil {
		return model.Nas{}, stacktrace.Propagate(err, "failed to find NAS by id")
	}

	return nas, nil
}

func (d *nasDomain) Update(ctx context.Context, id string, input model.NasInput) error {
	if id == "" {
		return stacktrace.NewError("id is empty")
	}

	// Re-encrypt password if changed
	if input.Password != "" {
		encryptedPassword, err := crypto.Encrypt(input.Password)
		if err != nil {
			return stacktrace.Propagate(err, "failed to encrypt password")
		}
		input.PasswordEncrypted = encryptedPassword
		input.Password = "" // Clear plain password
	}

	err := d.databasePort.Nas().Update(id, input)
	if err != nil {
		return stacktrace.Propagate(err, "failed to update NAS")
	}

	return nil
}

func (d *nasDomain) Delete(ctx context.Context, id string) error {
	if id == "" {
		return stacktrace.NewError("id is empty")
	}

	err := d.databasePort.Nas().Delete(id)
	if err != nil {
		return stacktrace.Propagate(err, "failed to delete NAS")
	}

	return nil
}

func (d *nasDomain) CountByTenantID(ctx context.Context, tenantID string) (int, error) {
	if tenantID == "" {
		return 0, stacktrace.NewError("tenantID is empty")
	}

	count, err := d.databasePort.Nas().CountByTenantID(tenantID)
	if err != nil {
		return 0, stacktrace.Propagate(err, "failed to count NAS by tenant ID")
	}

	return count, nil
}

func (d *nasDomain) SyncPackageWithProfileToNas(ctx context.Context, packageID string) error {
	if packageID == "" {
		return stacktrace.NewError("packageID is required")
	}

	// TODO: Re-implement when GetPPPoEProfiles is available
	// // Get existing profiles from MikroTik
	// existingProfiles, err := mikrotikPort.GetPPPoEProfiles(nas.RouterOSHost, nas.RouterOSPort, nasUsername, nasPassword, nas.UseSSL)
	// if err != nil {
	// 	return stacktrace.Propagate(err, "failed to get profiles from MikroTik")
	// }

	return stacktrace.NewError("SyncPackageWithProfileToNas not implemented - GetPPPoEProfiles method missing")
}

func (d *nasDomain) TestConnection(ctx context.Context, id string) error {
	if id == "" {
		return stacktrace.NewError("id is empty")
	}

	// Get NAS details
	nas, err := d.databasePort.Nas().FindByID(id)
	if err != nil {
		return stacktrace.Propagate(err, "failed to find NAS")
	}

	// Decrypt password
	password, err := crypto.Decrypt(nas.PasswordEncrypted)
	if err != nil {
		return stacktrace.Propagate(err, "failed to decrypt password")
	}

	// Test connection via MikroTik RouterOS API
	mikrotikPort := d.httpPort.Mikrotik()
	err = mikrotikPort.TestConnection(nas.Host, nas.ApiPort, nas.Username, password, nas.UseSSL)
	if err != nil {
		return stacktrace.Propagate(err, "failed to connect to MikroTik")
	}

	return nil
}

// provisionIsolirProfile provisions the isolir profile on a new NAS
func (d *nasDomain) provisionIsolirProfile(ctx context.Context, nas model.Nas) {
	// Get tenant settings
	tenantSettings, err := d.databasePort.TenantSetting().FindByTenantID(nas.TenantID)
	if err != nil {
		// No settings found, skip provisioning
		return
	}

	if tenantSettings.IsolirProfileName == "" {
		// No isolir profile configured
		return
	}

	// Decrypt NAS password
	password, err := crypto.Decrypt(nas.PasswordEncrypted)
	if err != nil {
		// Log error but don't fail NAS creation
		return
	}

	//  TODO: Check if isolir profile already exists on MikroTik
	// mikrotikPort := d.httpPort.Mikrotik()
	// profiles, err := mikrotikPort.GetPPPoEProfiles(nas.Host, nas.ApiPort, nas.Username, password, nas.UseSSL)
	// if err != nil {
	// 	// Log error but don't fail
	// 	return
	// }

	// // Check if isolir profile exists
	// profileExists := false
	// for _, profile := range profiles {
	// 	if profile["name"] == tenantSettings.IsolirProfileName {
	// 		profileExists = true
	// 		break
	// 	}
	// }

	// if !profileExists {
	// Assume profile doesn't exist, attempt to create it
	mikrotikPort := d.httpPort.Mikrotik()
	// Create isolir profile with limited bandwidth
	profileInput := model.MikrotikProfileInput{
		Name:      tenantSettings.IsolirProfileName,
		RateLimit: "128k/128k",
	}
	err = mikrotikPort.CreatePPPoEProfile(nas.Host, nas.ApiPort, nas.Username, password, nas.UseSSL, profileInput)
	if err != nil {
		// Log to mikrotik_sync_logs
		logInput := model.MikrotikSyncLogInput{
			TenantID:           nas.TenantID,
			NasID:              nas.ID,
			Action:             "provision_isolir_profile",
			ResourceType:       "pppoe_profile",
			ResourceIdentifier: tenantSettings.IsolirProfileName,
			Status:             "failed",
			ErrorMessage:       err.Error(),
		}
		_, _ = d.databasePort.MikrotikSyncLog().Create(logInput)
		return
	}

	// Log success
	logInput := model.MikrotikSyncLogInput{
		TenantID:           nas.TenantID,
		NasID:              nas.ID,
		Action:             "provision_isolir_profile",
		ResourceType:       "pppoe_profile",
		ResourceIdentifier: tenantSettings.IsolirProfileName,
		Status:             "success",
	}
	_, _ = d.databasePort.MikrotikSyncLog().Create(logInput)
}
