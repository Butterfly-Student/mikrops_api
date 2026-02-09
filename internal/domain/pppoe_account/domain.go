package pppoe_account

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/palantir/stacktrace"

	"mikrops/internal/model"
	outbound_port "mikrops/internal/port/outbound"
)

type PppoeAccountDomain interface {
	Create(ctx context.Context, input model.PppoeAccountInput) (model.PppoeAccount, error)
	FindByFilter(ctx context.Context, filter model.PppoeAccountFilter) ([]model.PppoeAccount, error)
	FindByID(ctx context.Context, id string) (model.PppoeAccount, error)
	Update(ctx context.Context, id string, input model.PppoeAccountInput) error
	Delete(ctx context.Context, id string) error
	Isolate(ctx context.Context, id string) (model.PppoeAccount, error)
	Restore(ctx context.Context, id string) (model.PppoeAccount, error)
}

type pppoeAccountDomain struct {
	databasePort outbound_port.DatabasePort
	messagePort  outbound_port.MessagePort
	cachePort    outbound_port.CachePort
	workflowPort outbound_port.WorkflowPort
}

func NewPppoeAccountDomain(
	databasePort outbound_port.DatabasePort,
	messagePort outbound_port.MessagePort,
	cachePort outbound_port.CachePort,
	workflowPort outbound_port.WorkflowPort,
) PppoeAccountDomain {
	return &pppoeAccountDomain{
		databasePort: databasePort,
		messagePort:  messagePort,
		cachePort:    cachePort,
		workflowPort: workflowPort,
	}
}

func (d *pppoeAccountDomain) Create(ctx context.Context, input model.PppoeAccountInput) (model.PppoeAccount, error) {
	if input.TenantID == "" {
		return model.PppoeAccount{}, stacktrace.NewError("tenant_id is required")
	}
	if input.CustomerID == "" {
		return model.PppoeAccount{}, stacktrace.NewError("customer_id is required")
	}
	if input.NasID == "" {
		return model.PppoeAccount{}, stacktrace.NewError("nas_id is required")
	}
	if input.PackageID == "" {
		return model.PppoeAccount{}, stacktrace.NewError("package_id is required")
	}

	// Generate username if not provided
	if input.Username == "" {
		input.Username = generatePPPoEUsername()
	}

	// Generate password if not provided
	if input.Password != "" {
		input.PasswordEncrypted = input.Password
		input.Password = ""
	} else if input.PasswordEncrypted == "" {
		input.PasswordEncrypted = generatePPPoEPassword()
	}

	// Set profile name from package if not provided
	if input.ProfileName == "" {
		pkg, err := d.databasePort.InternetPackage().FindByID(input.PackageID)
		if err == nil && pkg.ProfileName != "" {
			input.ProfileName = pkg.ProfileName
		}
	}

	input.Status = model.PppoeStatusActive
	input.SyncStatus = model.PppoeSyncStatusPending

	account, err := d.databasePort.PppoeAccount().Create(input)
	if err != nil {
		return model.PppoeAccount{}, stacktrace.Propagate(err, "failed to create pppoe account")
	}

	return account, nil
}

func (d *pppoeAccountDomain) FindByFilter(ctx context.Context, filter model.PppoeAccountFilter) ([]model.PppoeAccount, error) {
	if filter.IsEmpty() {
		return nil, stacktrace.NewError("filter is empty")
	}

	accounts, err := d.databasePort.PppoeAccount().FindByFilter(filter)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to find pppoe accounts")
	}

	return accounts, nil
}

func (d *pppoeAccountDomain) FindByID(ctx context.Context, id string) (model.PppoeAccount, error) {
	if id == "" {
		return model.PppoeAccount{}, stacktrace.NewError("id is empty")
	}

	account, err := d.databasePort.PppoeAccount().FindByID(id)
	if err != nil {
		return model.PppoeAccount{}, stacktrace.Propagate(err, "failed to find pppoe account")
	}

	return account, nil
}

func (d *pppoeAccountDomain) Update(ctx context.Context, id string, input model.PppoeAccountInput) error {
	if id == "" {
		return stacktrace.NewError("id is empty")
	}

	if input.Password != "" {
		input.PasswordEncrypted = input.Password
		input.Password = ""
	}

	err := d.databasePort.PppoeAccount().Update(id, input)
	if err != nil {
		return stacktrace.Propagate(err, "failed to update pppoe account")
	}

	return nil
}

func (d *pppoeAccountDomain) Delete(ctx context.Context, id string) error {
	if id == "" {
		return stacktrace.NewError("id is empty")
	}

	err := d.databasePort.PppoeAccount().Delete(id)
	if err != nil {
		return stacktrace.Propagate(err, "failed to delete pppoe account")
	}

	return nil
}

func (d *pppoeAccountDomain) Isolate(ctx context.Context, id string) (model.PppoeAccount, error) {
	if id == "" {
		return model.PppoeAccount{}, stacktrace.NewError("id is empty")
	}

	account, err := d.databasePort.PppoeAccount().FindByID(id)
	if err != nil {
		return model.PppoeAccount{}, stacktrace.Propagate(err, "failed to find pppoe account")
	}

	if account.Status != model.PppoeStatusActive {
		return model.PppoeAccount{}, stacktrace.NewError("account is not in active status")
	}

	// Get isolir profile name from tenant settings
	isolirProfile := "ISOLIR_MIKROPS"
	tenantSetting, err := d.databasePort.TenantSetting().FindByTenantID(account.TenantID)
	if err == nil && tenantSetting.IsolirProfileName != "" {
		isolirProfile = tenantSetting.IsolirProfileName
	}

	now := time.Now()
	updateInput := model.PppoeAccountInput{
		OriginalProfile: account.ProfileName,
		ProfileName:     isolirProfile,
		Status:          model.PppoeStatusIsolated,
		SyncStatus:      model.PppoeSyncStatusPending,
		LastSyncAt:      &now,
	}

	err = d.databasePort.PppoeAccount().Update(id, updateInput)
	if err != nil {
		return model.PppoeAccount{}, stacktrace.Propagate(err, "failed to isolate pppoe account")
	}

	updatedAccount, err := d.databasePort.PppoeAccount().FindByID(id)
	if err != nil {
		return model.PppoeAccount{}, stacktrace.Propagate(err, "failed to fetch updated account")
	}

	return updatedAccount, nil
}

func (d *pppoeAccountDomain) Restore(ctx context.Context, id string) (model.PppoeAccount, error) {
	if id == "" {
		return model.PppoeAccount{}, stacktrace.NewError("id is empty")
	}

	account, err := d.databasePort.PppoeAccount().FindByID(id)
	if err != nil {
		return model.PppoeAccount{}, stacktrace.Propagate(err, "failed to find pppoe account")
	}

	if account.Status != model.PppoeStatusIsolated {
		return model.PppoeAccount{}, stacktrace.NewError("account is not in isolated status")
	}

	profileToRestore := account.OriginalProfile
	if profileToRestore == "" {
		profileToRestore = account.ProfileName
	}

	now := time.Now()
	updateInput := model.PppoeAccountInput{
		ProfileName:     profileToRestore,
		OriginalProfile: "",
		Status:          model.PppoeStatusActive,
		SyncStatus:      model.PppoeSyncStatusPending,
		LastSyncAt:      &now,
	}

	err = d.databasePort.PppoeAccount().Update(id, updateInput)
	if err != nil {
		return model.PppoeAccount{}, stacktrace.Propagate(err, "failed to restore pppoe account")
	}

	updatedAccount, err := d.databasePort.PppoeAccount().FindByID(id)
	if err != nil {
		return model.PppoeAccount{}, stacktrace.Propagate(err, "failed to fetch updated account")
	}

	return updatedAccount, nil
}

func generatePPPoEUsername() string {
	b := make([]byte, 4)
	rand.Read(b)
	return fmt.Sprintf("pppoe_%s", hex.EncodeToString(b))
}

func generatePPPoEPassword() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}
