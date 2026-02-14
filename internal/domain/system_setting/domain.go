package system_setting

import (
	"context"

	"github.com/google/uuid"
	"github.com/palantir/stacktrace"

	"go-template/internal/model"
	outbound_port "go-template/internal/port/outbound"
)

type SystemSettingDomain interface {
	GetByKey(ctx context.Context, key string) (*model.SystemSetting, error)
	GetAll(ctx context.Context, category string) ([]model.SystemSetting, error)
	UpdateByKey(ctx context.Context, key string, value string) error
	Create(ctx context.Context, input model.SystemSettingInput) (*model.SystemSetting, error)
	Update(ctx context.Context, id string, input model.SystemSettingInput) (*model.SystemSetting, error)
	GetPublicSettings(ctx context.Context) ([]model.SystemSetting, error)
}

type systemSettingDomain struct {
	databasePort outbound_port.DatabasePort
}

func NewSystemSettingDomain(
	databasePort outbound_port.DatabasePort,
) SystemSettingDomain {
	return &systemSettingDomain{
		databasePort: databasePort,
	}
}

func (d *systemSettingDomain) GetByKey(ctx context.Context, key string) (*model.SystemSetting, error) {
	if key == "" {
		return nil, stacktrace.NewError("key is required")
	}

	databaseSystemSettingPort := d.databasePort.SystemSetting()
	setting, err := databaseSystemSettingPort.GetByKey(key)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get system setting")
	}

	if setting == nil {
		return nil, stacktrace.NewError("system setting not found")
	}

	return setting, nil
}

func (d *systemSettingDomain) GetAll(ctx context.Context, category string) ([]model.SystemSetting, error) {
	databaseSystemSettingPort := d.databasePort.SystemSetting()
	settings, err := databaseSystemSettingPort.GetAll(category)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get system settings")
	}

	return settings, nil
}

func (d *systemSettingDomain) UpdateByKey(ctx context.Context, key string, value string) error {
	if key == "" {
		return stacktrace.NewError("key is required")
	}

	// Check if setting exists
	setting, err := d.GetByKey(ctx, key)
	if err != nil {
		return stacktrace.Propagate(err, "failed to find system setting")
	}

	if setting == nil {
		return stacktrace.NewError("system setting not found")
	}

	databaseSystemSettingPort := d.databasePort.SystemSetting()
	err = databaseSystemSettingPort.UpdateByKey(key, value)
	if err != nil {
		return stacktrace.Propagate(err, "failed to update system setting")
	}

	return nil
}

func (d *systemSettingDomain) Create(ctx context.Context, input model.SystemSettingInput) (*model.SystemSetting, error) {
	// Validate input
	if input.Key == "" {
		return nil, stacktrace.NewError("key is required")
	}
	if input.Category == "" {
		return nil, stacktrace.NewError("category is required")
	}

	// Check if key already exists
	databaseSystemSettingPort := d.databasePort.SystemSetting()
	existing, err := databaseSystemSettingPort.GetByKey(input.Key)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to check existing setting")
	}
	if existing != nil {
		return nil, stacktrace.NewError("setting key already exists")
	}

	// Create setting
	setting := &model.SystemSetting{
		ID:          uuid.New(),
		Key:         input.Key,
		Value:       input.Value,
		ValueType:   input.ValueType,
		Category:    input.Category,
		Description: input.Description,
		IsPublic:    input.IsPublic,
	}

	err = databaseSystemSettingPort.Create(setting)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to create system setting")
	}

	return setting, nil
}

func (d *systemSettingDomain) Update(ctx context.Context, id string, input model.SystemSettingInput) (*model.SystemSetting, error) {
	if id == "" {
		return nil, stacktrace.NewError("id is required")
	}

	settingID, err := uuid.Parse(id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "invalid uuid format")
	}

	// Find existing setting (we need to query by ID, so we'll get all and filter)
	databaseSystemSettingPort := d.databasePort.SystemSetting()
	allSettings, err := databaseSystemSettingPort.GetAll("")
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to find system setting")
	}

	var setting *model.SystemSetting
	for i := range allSettings {
		if allSettings[i].ID == settingID {
			setting = &allSettings[i]
			break
		}
	}

	if setting == nil {
		return nil, stacktrace.NewError("system setting not found")
	}

	// Check if key is being changed and already exists
	if input.Key != setting.Key {
		existing, err := databaseSystemSettingPort.GetByKey(input.Key)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to check existing setting")
		}
		if existing != nil {
			return nil, stacktrace.NewError("setting key already exists")
		}
	}

	// Update setting fields
	setting.Key = input.Key
	setting.Value = input.Value
	setting.ValueType = input.ValueType
	setting.Category = input.Category
	setting.Description = input.Description
	setting.IsPublic = input.IsPublic

	err = databaseSystemSettingPort.Update(setting)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to update system setting")
	}

	return setting, nil
}

func (d *systemSettingDomain) GetPublicSettings(ctx context.Context) ([]model.SystemSetting, error) {
	databaseSystemSettingPort := d.databasePort.SystemSetting()
	allSettings, err := databaseSystemSettingPort.GetAll("")
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get system settings")
	}

	// Filter public settings
	var publicSettings []model.SystemSetting
	for _, setting := range allSettings {
		if setting.IsPublic {
			publicSettings = append(publicSettings, setting)
		}
	}

	return publicSettings, nil
}
