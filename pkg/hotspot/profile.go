package hotspot

import (
	"context"
	"fmt"

	"github.com/go-routeros/routeros/v3/proto"
	"go-template/pkg/hotspot/internal"
)

// CreateProfile creates new hotspot user profile on RouterOS
func (c *hotspotClient) CreateProfile(ctx context.Context, profile *Profile) error {
	// Validate
	if profile.Name == "" {
		return NewError("create profile", fmt.Errorf("profile name is required"))
	}

	// Build on-login script
	onLoginScript := internal.BuildOnLoginScript(
		profile.ExpiryMode,
		profile.Price,
		profile.SellingPrice,
		profile.Validity,
		profile.LockUser,
	)

	_, err := c.execute(ctx, PathHotspotUserProfile+"/add",
		"=name="+profile.Name,
		"=shared-users="+fmt.Sprintf("%d", profile.SharedUsers),
		"=rate-limit="+profile.RateLimit,
		"=on-login="+onLoginScript,
		"=keepalive-timeout="+profile.KeepaliveTimeout,
	)

	if err != nil {
		return WrapError("create profile", err)
	}

	return nil
}

// GetProfile retrieves profile from RouterOS by name
func (c *hotspotClient) GetProfile(ctx context.Context, name string) (*Profile, error) {
	if name == "" {
		return nil, NewError("get profile", fmt.Errorf("profile name is required"))
	}

	reply, err := c.execute(ctx, PathHotspotUserProfile+"/print", "?name="+name)
	if err != nil {
		return nil, WrapError("get profile", err)
	}

	if len(reply.Re) == 0 {
		return nil, ErrProfileNotFound
	}

	return c.mapReplyToProfile(reply.Re[0]), nil
}

// GetAllProfiles retrieves all profiles from RouterOS
func (c *hotspotClient) GetAllProfiles(ctx context.Context) ([]Profile, error) {
	reply, err := c.execute(ctx, PathHotspotUserProfile+"/print")
	if err != nil {
		return nil, WrapError("get all profiles", err)
	}

	profiles := make([]Profile, 0, len(reply.Re))
	for _, re := range reply.Re {
		profiles = append(profiles, *c.mapReplyToProfile(re))
	}

	return profiles, nil
}

// UpdateProfile updates existing profile on RouterOS
func (c *hotspotClient) UpdateProfile(ctx context.Context, name string, updates *ProfileUpdate) error {
	if name == "" {
		return NewError("update profile", fmt.Errorf("profile name is required"))
	}
	if updates == nil {
		return nil
	}

	// Get current profile
	current, err := c.GetProfile(ctx, name)
	if err != nil {
		if err == ErrProfileNotFound {
			return ErrProfileNotFound
		}
		return WrapError("update profile", err)
	}

	// Apply updates
	if updates.RateLimit != nil {
		current.RateLimit = *updates.RateLimit
	}
	if updates.SharedUsers != nil {
		current.SharedUsers = *updates.SharedUsers
	}
	if updates.Validity != nil {
		current.Validity = *updates.Validity
	}
	if updates.Price != nil {
		current.Price = *updates.Price
	}
	if updates.SellingPrice != nil {
		current.SellingPrice = *updates.SellingPrice
	}
	if updates.ExpiryMode != nil {
		current.ExpiryMode = *updates.ExpiryMode
	}
	if updates.LockUser != nil {
		current.LockUser = *updates.LockUser
	}
	if updates.KeepaliveTimeout != nil {
		current.KeepaliveTimeout = *updates.KeepaliveTimeout
	}

	// Find profile ID
	profileID, err := c.findResourceID(ctx, PathHotspotUserProfile, "name", name)
	if err != nil {
		return WrapError("update profile", err)
	}

	// Rebuild on-login script
	onLoginScript := internal.BuildOnLoginScript(
		current.ExpiryMode,
		current.Price,
		current.SellingPrice,
		current.Validity,
		current.LockUser,
	)

	_, err = c.execute(ctx, PathHotspotUserProfile+"/set",
		"=.id="+profileID,
		"=rate-limit="+current.RateLimit,
		"=shared-users="+fmt.Sprintf("%d", current.SharedUsers),
		"=on-login="+onLoginScript,
		"=keepalive-timeout="+current.KeepaliveTimeout,
	)

	if err != nil {
		return WrapError("update profile", err)
	}

	return nil
}

// DeleteProfile deletes profile from RouterOS
func (c *hotspotClient) DeleteProfile(ctx context.Context, name string) error {
	if name == "" {
		return NewError("delete profile", fmt.Errorf("profile name is required"))
	}

	profileID, err := c.findResourceID(ctx, PathHotspotUserProfile, "name", name)
	if err != nil {
		if err == ErrProfileNotFound {
			return ErrProfileNotFound
		}
		return WrapError("delete profile", err)
	}

	_, err = c.execute(ctx, PathHotspotUserProfile+"/remove", "=.id="+profileID)
	if err != nil {
		return WrapError("delete profile", err)
	}

	return nil
}

// SyncProfileToRouter ensures profile exists on RouterOS with current settings
func (c *hotspotClient) SyncProfileToRouter(ctx context.Context, profile *Profile) error {
	_, err := c.GetProfile(ctx, profile.Name)
	if err != nil {
		if err == ErrProfileNotFound {
			// Create new profile
			return c.CreateProfile(ctx, profile)
		}
		return err
	}

	// Update existing profile
	update := &ProfileUpdate{
		RateLimit:        &profile.RateLimit,
		SharedUsers:      &profile.SharedUsers,
		Validity:         &profile.Validity,
		Price:            &profile.Price,
		SellingPrice:     &profile.SellingPrice,
		ExpiryMode:       &profile.ExpiryMode,
		LockUser:         &profile.LockUser,
		KeepaliveTimeout: &profile.KeepaliveTimeout,
	}
	return c.UpdateProfile(ctx, profile.Name, update)
}

// GetProfileSettings extracts price/validity from on-login script
func (c *hotspotClient) GetProfileSettings(ctx context.Context, name string) (*Profile, error) {
	return c.GetProfile(ctx, name)
}

func (c *hotspotClient) mapReplyToProfile(re *proto.Sentence) *Profile {
	profile := &Profile{
		Name:             re.Map["name"],
		SharedUsers:      internal.ParseReplyToInt(re.Map["shared-users"]),
		RateLimit:        re.Map["rate-limit"],
		KeepaliveTimeout: re.Map["keepalive-timeout"],
		OnLoginScript:    re.Map["on-login"],
	}

	// Parse on-login script
	parsed, _ := internal.ParseOnLoginScript(re.Map["on-login"])
	if parsed != nil {
		profile.ExpiryMode = parsed.ExpiryMode
		profile.Price = parsed.Price
		profile.Validity = parsed.Validity
		profile.SellingPrice = parsed.SellingPrice
		profile.LockUser = parsed.LockUser
	}

	return profile
}
