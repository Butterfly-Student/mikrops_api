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
	GetIdentity(ctx context.Context, id string) (string, error)
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

func (d *nasDomain) GetIdentity(ctx context.Context, id string) (string, error) {
	if id == "" {
		return "", stacktrace.NewError("id is empty")
	}

	nas, err := d.databasePort.Nas().FindByID(id)
	if err != nil {
		return "", stacktrace.Propagate(err, "failed to find NAS")
	}

	password, err := crypto.Decrypt(nas.PasswordEncrypted)
	if err != nil {
		return "", stacktrace.Propagate(err, "failed to decrypt password")
	}

	mikrotikPort := d.httpPort.Mikrotik()
	identity, err := mikrotikPort.GetIdentity(nas.Host, nas.ApiPort, nas.Username, password, nas.UseSSL)
	if err != nil {
		return "", stacktrace.Propagate(err, "failed to get identity from MikroTik")
	}

	return identity, nil
}
