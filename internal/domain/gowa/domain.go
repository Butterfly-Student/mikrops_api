package gowa_domain

import (
	"context"
	"fmt"

	outbound_port "go-template/internal/port/outbound"
	"go-template/pkg/gowa"
)

// GowaDomain defines business operations for WhatsApp gateway management via Gowa API.
type GowaDomain interface {
	// Device management
	ListDevices(ctx context.Context) ([]gowa.DeviceInfo, error)
	AddDevice(ctx context.Context, deviceID string) (*gowa.DeviceInfo, error)
	GetDevice(ctx context.Context, deviceID string) (*gowa.DeviceInfo, error)
	RemoveDevice(ctx context.Context, deviceID string) error
	LoginDevice(ctx context.Context, deviceID string) (*gowa.LoginResponse, error)
	LoginDeviceWithCode(ctx context.Context, deviceID, phone string) (*gowa.LoginWithCodeResponse, error)
	LogoutDevice(ctx context.Context, deviceID string) error
	ReconnectDevice(ctx context.Context, deviceID string) error
	GetDeviceStatus(ctx context.Context, deviceID string) (*gowa.DeviceStatusResult, error)

	// App / session (uses default or specified device)
	AppLogin(ctx context.Context, deviceID string) (*gowa.LoginResponse, error)
	AppLoginWithCode(ctx context.Context, phone, deviceID string) (*gowa.LoginWithCodeResponse, error)
	AppLogout(ctx context.Context, deviceID string) error
	AppReconnect(ctx context.Context, deviceID string) error
	AppStatus(ctx context.Context, deviceID string) (*gowa.AppStatusResponse, error)

	// Group management
	GetMyGroups(ctx context.Context, deviceID string) ([]gowa.Group, error)
	FindGroupByName(ctx context.Context, name, deviceID string) (*gowa.Group, error)
	GetGroupInfo(ctx context.Context, groupJID, deviceID string) (*gowa.GroupInfoResponse, error)
	GetGroupInviteLink(ctx context.Context, groupJID, deviceID string) (*gowa.GetGroupInviteLinkResponse, error)

	// User information
	CheckUser(ctx context.Context, phone, deviceID string) (bool, error)
	GetUserInfo(ctx context.Context, phone, deviceID string) (*gowa.UserInfoResponse, error)
	GetMyContacts(ctx context.Context, deviceID string) ([]map[string]interface{}, error)

	// Send messages
	SendMessage(ctx context.Context, req gowa.SendMessageRequest, deviceID string) (*gowa.SendResponse, error)
	SendTextMessage(ctx context.Context, phone, message, deviceID string) (*gowa.SendResponse, error)
	SendImageFromURL(ctx context.Context, phone, imageURL, caption, deviceID string) (*gowa.SendResponse, error)
	SendFileFromURL(ctx context.Context, phone, fileURL, caption, deviceID string) (*gowa.SendResponse, error)
	SendVideoFromURL(ctx context.Context, phone, videoURL, caption, deviceID string) (*gowa.SendResponse, error)
}

type domain struct {
	gowaPort outbound_port.GowaPort
}

func NewGowaDomain(gowaPort outbound_port.GowaPort) GowaDomain {
	return &domain{gowaPort: gowaPort}
}

func (d *domain) client(deviceID string) (*gowa.Client, error) {
	c, err := d.gowaPort.GetClient()
	if err != nil {
		return nil, fmt.Errorf("whatsapp gateway not available: %w", err)
	}
	if deviceID != "" {
		return c.WithDeviceID(deviceID), nil
	}
	return c, nil
}

// ─── Device Management ────────────────────────────────────────────────────────

func (d *domain) ListDevices(ctx context.Context) ([]gowa.DeviceInfo, error) {
	c, err := d.client("")
	if err != nil {
		return nil, err
	}
	return c.ListDevices(ctx)
}

func (d *domain) AddDevice(ctx context.Context, deviceID string) (*gowa.DeviceInfo, error) {
	c, err := d.client("")
	if err != nil {
		return nil, err
	}
	return c.AddDevice(ctx, deviceID)
}

func (d *domain) GetDevice(ctx context.Context, deviceID string) (*gowa.DeviceInfo, error) {
	c, err := d.client("")
	if err != nil {
		return nil, err
	}
	return c.GetDevice(ctx, deviceID)
}

func (d *domain) RemoveDevice(ctx context.Context, deviceID string) error {
	c, err := d.client("")
	if err != nil {
		return err
	}
	_, err = c.RemoveDevice(ctx, deviceID)
	return err
}

func (d *domain) LoginDevice(ctx context.Context, deviceID string) (*gowa.LoginResponse, error) {
	c, err := d.client("")
	if err != nil {
		return nil, err
	}
	return c.LoginDevice(ctx, deviceID)
}

func (d *domain) LoginDeviceWithCode(ctx context.Context, deviceID, phone string) (*gowa.LoginWithCodeResponse, error) {
	c, err := d.client("")
	if err != nil {
		return nil, err
	}
	return c.LoginDeviceWithCode(ctx, deviceID, phone)
}

func (d *domain) LogoutDevice(ctx context.Context, deviceID string) error {
	c, err := d.client("")
	if err != nil {
		return err
	}
	_, err = c.LogoutDevice(ctx, deviceID)
	return err
}

func (d *domain) ReconnectDevice(ctx context.Context, deviceID string) error {
	c, err := d.client("")
	if err != nil {
		return err
	}
	_, err = c.ReconnectDevice(ctx, deviceID)
	return err
}

func (d *domain) GetDeviceStatus(ctx context.Context, deviceID string) (*gowa.DeviceStatusResult, error) {
	c, err := d.client("")
	if err != nil {
		return nil, err
	}
	return c.GetDeviceStatus(ctx, deviceID)
}

// ─── App / Session ────────────────────────────────────────────────────────────

func (d *domain) AppLogin(ctx context.Context, deviceID string) (*gowa.LoginResponse, error) {
	c, err := d.client(deviceID)
	if err != nil {
		return nil, err
	}
	return c.Login(ctx)
}

func (d *domain) AppLoginWithCode(ctx context.Context, phone, deviceID string) (*gowa.LoginWithCodeResponse, error) {
	c, err := d.client(deviceID)
	if err != nil {
		return nil, err
	}
	return c.LoginWithCode(ctx, phone)
}

func (d *domain) AppLogout(ctx context.Context, deviceID string) error {
	c, err := d.client(deviceID)
	if err != nil {
		return err
	}
	_, err = c.Logout(ctx)
	return err
}

func (d *domain) AppReconnect(ctx context.Context, deviceID string) error {
	c, err := d.client(deviceID)
	if err != nil {
		return err
	}
	_, err = c.Reconnect(ctx)
	return err
}

func (d *domain) AppStatus(ctx context.Context, deviceID string) (*gowa.AppStatusResponse, error) {
	c, err := d.client(deviceID)
	if err != nil {
		return nil, err
	}
	return c.GetStatus(ctx)
}

// ─── Group Management ─────────────────────────────────────────────────────────

func (d *domain) GetMyGroups(ctx context.Context, deviceID string) ([]gowa.Group, error) {
	c, err := d.client(deviceID)
	if err != nil {
		return nil, err
	}
	return c.GetMyGroups(ctx)
}

func (d *domain) FindGroupByName(ctx context.Context, name, deviceID string) (*gowa.Group, error) {
	c, err := d.client(deviceID)
	if err != nil {
		return nil, err
	}
	return c.FindGroupByName(ctx, name)
}

func (d *domain) GetGroupInfo(ctx context.Context, groupJID, deviceID string) (*gowa.GroupInfoResponse, error) {
	c, err := d.client(deviceID)
	if err != nil {
		return nil, err
	}
	return c.GetGroupInfo(ctx, groupJID)
}

func (d *domain) GetGroupInviteLink(ctx context.Context, groupJID, deviceID string) (*gowa.GetGroupInviteLinkResponse, error) {
	c, err := d.client(deviceID)
	if err != nil {
		return nil, err
	}
	return c.GetGroupInviteLink(ctx, groupJID)
}

// ─── User Information ─────────────────────────────────────────────────────────

func (d *domain) CheckUser(ctx context.Context, phone, deviceID string) (bool, error) {
	c, err := d.client(deviceID)
	if err != nil {
		return false, err
	}
	return c.CheckUser(ctx, phone)
}

func (d *domain) GetUserInfo(ctx context.Context, phone, deviceID string) (*gowa.UserInfoResponse, error) {
	c, err := d.client(deviceID)
	if err != nil {
		return nil, err
	}
	return c.GetUserInfo(ctx, phone)
}

func (d *domain) GetMyContacts(ctx context.Context, deviceID string) ([]map[string]interface{}, error) {
	c, err := d.client(deviceID)
	if err != nil {
		return nil, err
	}
	return c.GetMyContacts(ctx)
}

// ─── Send Messages ────────────────────────────────────────────────────────────

func (d *domain) SendTextMessage(ctx context.Context, phone, message, deviceID string) (*gowa.SendResponse, error) {
	return d.SendMessage(ctx, gowa.SendMessageRequest{Phone: phone, Message: message}, deviceID)
}

func (d *domain) SendMessage(ctx context.Context, req gowa.SendMessageRequest, deviceID string) (*gowa.SendResponse, error) {
	c, err := d.client(deviceID)
	if err != nil {
		return nil, err
	}
	return c.SendMessageWithDevice(ctx, req, "")
}

func (d *domain) SendImageFromURL(ctx context.Context, phone, imageURL, caption, deviceID string) (*gowa.SendResponse, error) {
	c, err := d.client(deviceID)
	if err != nil {
		return nil, err
	}
	return c.SendImageFromURL(ctx, phone, imageURL, caption, "")
}

func (d *domain) SendFileFromURL(ctx context.Context, phone, fileURL, caption, deviceID string) (*gowa.SendResponse, error) {
	c, err := d.client(deviceID)
	if err != nil {
		return nil, err
	}
	return c.SendFileFromURL(ctx, phone, fileURL, caption, "")
}

func (d *domain) SendVideoFromURL(ctx context.Context, phone, videoURL, caption, deviceID string) (*gowa.SendResponse, error) {
	c, err := d.client(deviceID)
	if err != nil {
		return nil, err
	}
	return c.SendVideoFromURL(ctx, phone, videoURL, caption, "")
}
