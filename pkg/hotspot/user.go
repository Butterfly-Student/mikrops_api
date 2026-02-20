package hotspot

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-routeros/routeros/v3/proto"
	"go-template/pkg/hotspot/internal"
)

// CreateUser creates a new hotspot user on RouterOS.
// The user will be assigned to the specified profile and can have
// time and data limits configured.
func (c *hotspotClient) CreateUser(ctx context.Context, user *User) error {
	// Validate input
	if user.Name == "" {
		return NewError("create user", fmt.Errorf("username is required"))
	}
	if user.Password == "" {
		return NewError("create user", fmt.Errorf("password is required"))
	}
	if user.Profile == "" {
		return NewError("create user", fmt.Errorf("profile is required"))
	}

	// Build default comment if not provided.
	// Note: expiryDate defaults to today — callers should set Comment explicitly
	// (e.g. via GenerateVouchers which calculates expiry from the profile validity).
	if user.Comment == "" {
		mode := internal.GetUserMode(user.Name, user.Password)
		prefix := internal.ExtractPrefix(user.Name)
		expiryDate := time.Now().Format("Jan/02/2006")
		user.Comment = internal.BuildUserComment(expiryDate, mode, prefix)
	}

	// Set default server
	if user.Server == "" {
		user.Server = DefaultServer
	}

	// Build command arguments
	args := []string{
		"=server=" + user.Server,
		"=name=" + user.Name,
		"=password=" + user.Password,
		"=profile=" + user.Profile,
		"=disabled=" + internal.BoolToString(user.Disabled),
		"=limit-uptime=" + fmt.Sprintf("%d", user.LimitUptime),
		"=limit-bytes-total=" + fmt.Sprintf("%d", user.LimitBytesTotal),
		"=comment=" + user.Comment,
	}

	_, err := c.execute(ctx, PathHotspotUser+"/add", args...)
	if err != nil {
		return WrapError("create user", err)
	}

	return nil
}

// GetUser retrieves a user from RouterOS by username
func (c *hotspotClient) GetUser(ctx context.Context, username string) (*User, error) {
	if username == "" {
		return nil, NewError("get user", fmt.Errorf("username is required"))
	}

	reply, err := c.execute(ctx, PathHotspotUser+"/print", "?name="+username)
	if err != nil {
		return nil, WrapError("get user", err)
	}

	if len(reply.Re) == 0 {
		return nil, ErrUserNotFound
	}

	return c.mapReplyToUser(reply.Re[0]), nil
}

// GetAllUsers retrieves all users from RouterOS, optionally filtered
func (c *hotspotClient) GetAllUsers(ctx context.Context, filter *UserFilter) ([]User, error) {
	args := c.buildUserFilterArgs(filter)

	reply, err := c.execute(ctx, PathHotspotUser+"/print", args...)
	if err != nil {
		return nil, WrapError("get all users", err)
	}

	users := make([]User, 0, len(reply.Re))
	for _, re := range reply.Re {
		users = append(users, *c.mapReplyToUser(re))
	}

	// Apply pagination
	if filter != nil && filter.Limit > 0 {
		users = c.applyPagination(users, filter.Offset, filter.Limit)
	}

	return users, nil
}

// GetUsersByProfile retrieves users assigned to a specific profile
func (c *hotspotClient) GetUsersByProfile(ctx context.Context, profile string) ([]User, error) {
	if profile == "" {
		return nil, ErrInvalidProfile
	}
	return c.GetAllUsers(ctx, &UserFilter{Profile: profile})
}

// GetUsersByComment retrieves users whose comment contains the given substring.
// RouterOS ?comment= filter is an exact match, so we fetch all users and
// perform a client-side substring search to support partial comment filters
// (e.g. filtering by "vc-CAFE" to find all vouchers for a given prefix).
func (c *hotspotClient) GetUsersByComment(ctx context.Context, comment string) ([]User, error) {
	if comment == "" {
		return nil, fmt.Errorf("comment filter cannot be empty")
	}

	all, err := c.GetAllUsers(ctx, nil)
	if err != nil {
		return nil, WrapError("get users by comment", err)
	}

	filtered := make([]User, 0)
	for _, u := range all {
		if strings.Contains(u.Comment, comment) {
			filtered = append(filtered, u)
		}
	}
	return filtered, nil
}

// UpdateUser updates an existing user's properties
func (c *hotspotClient) UpdateUser(ctx context.Context, username string, updates *UserUpdate) error {
	if username == "" {
		return NewError("update user", fmt.Errorf("username is required"))
	}
	if updates == nil {
		return nil
	}

	// Find user ID
	userID, err := c.findResourceID(ctx, PathHotspotUser, "name", username)
	if err != nil {
		if err == ErrUserNotFound {
			return ErrUserNotFound
		}
		return WrapError("update user", err)
	}

	// Build update arguments
	args := []string{"=.id=" + userID}

	if updates.Profile != nil {
		args = append(args, "=profile="+*updates.Profile)
	}
	if updates.Disabled != nil {
		args = append(args, "=disabled="+internal.BoolToString(*updates.Disabled))
	}
	if updates.Comment != nil {
		args = append(args, "=comment="+*updates.Comment)
	}
	if updates.LimitUptime != nil {
		args = append(args, "=limit-uptime="+fmt.Sprintf("%d", *updates.LimitUptime))
	}
	if updates.LimitBytesTotal != nil {
		args = append(args, "=limit-bytes-total="+fmt.Sprintf("%d", *updates.LimitBytesTotal))
	}
	if updates.LimitBytesIn != nil {
		args = append(args, "=limit-bytes-in="+fmt.Sprintf("%d", *updates.LimitBytesIn))
	}
	if updates.LimitBytesOut != nil {
		args = append(args, "=limit-bytes-out="+fmt.Sprintf("%d", *updates.LimitBytesOut))
	}

	_, err = c.execute(ctx, PathHotspotUser+"/set", args...)
	if err != nil {
		return WrapError("update user", err)
	}

	return nil
}

// DeleteUser removes a user from RouterOS
func (c *hotspotClient) DeleteUser(ctx context.Context, username string) error {
	if username == "" {
		return NewError("delete user", fmt.Errorf("username is required"))
	}

	userID, err := c.findResourceID(ctx, PathHotspotUser, "name", username)
	if err != nil {
		if err == ErrUserNotFound {
			return ErrUserNotFound
		}
		return WrapError("delete user", err)
	}

	_, err = c.execute(ctx, PathHotspotUser+"/remove", "=.id="+userID)
	if err != nil {
		return WrapError("delete user", err)
	}

	return nil
}

// DisableUser disables a user account
func (c *hotspotClient) DisableUser(ctx context.Context, username string) error {
	return c.UpdateUser(ctx, username, &UserUpdate{
		Disabled: boolPtr(true),
	})
}

// EnableUser enables a previously disabled user account
func (c *hotspotClient) EnableUser(ctx context.Context, username string) error {
	return c.UpdateUser(ctx, username, &UserUpdate{
		Disabled: boolPtr(false),
	})
}

// RemoveExpiredUsers removes users whose expiry date in comment has passed
// Returns the number of users removed
func (c *hotspotClient) RemoveExpiredUsers(ctx context.Context, profile string) (int, error) {
	users, err := c.GetAllUsers(ctx, &UserFilter{Profile: profile})
	if err != nil {
		return 0, err
	}

	removedCount := 0
	now := time.Now()

	for _, user := range users {
		expiryInfo, err := internal.ParseUserComment(user.Comment)
		if err != nil {
			continue // Skip users with invalid comment format
		}

		if expiryInfo.ExpiryDate.Before(now) {
			err := c.DeleteUser(ctx, user.Name)
			if err == nil {
				removedCount++
			}
		}
	}

	return removedCount, nil
}

// RemoveUnusedVouchers removes vouchers that have never been used (uptime=0s)
// Returns the number of users removed
func (c *hotspotClient) RemoveUnusedVouchers(ctx context.Context, profile string) (int, error) {
	users, err := c.GetAllUsers(ctx, &UserFilter{Profile: profile})
	if err != nil {
		return 0, err
	}

	removedCount := 0
	for _, user := range users {
		// Check if voucher mode and never used
		if user.Uptime == "0s" || user.Uptime == "" {
			err := c.DeleteUser(ctx, user.Name)
			if err == nil {
				removedCount++
			}
		}
	}

	return removedCount, nil
}

// BatchCreateUsers creates multiple users in batch
// Returns a VoucherResult with success/failure counts and any errors
func (c *hotspotClient) BatchCreateUsers(ctx context.Context, users []User) (*VoucherResult, error) {
	result := &VoucherResult{
		Vouchers: make([]User, 0),
		Errors:   make([]string, 0),
	}

	for _, user := range users {
		err := c.CreateUser(ctx, &user)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors,
				fmt.Sprintf("%s: %v", user.Name, err))
		} else {
			result.Success++
			result.Vouchers = append(result.Vouchers, user)
		}
	}

	return result, nil
}

// BatchRemoveUsers removes multiple users by username
// Returns the number of successfully removed users
func (c *hotspotClient) BatchRemoveUsers(ctx context.Context, usernames []string) (int, error) {
	removedCount := 0
	for _, username := range usernames {
		err := c.DeleteUser(ctx, username)
		if err == nil {
			removedCount++
		}
	}
	return removedCount, nil
}

// Helper methods

func (c *hotspotClient) mapReplyToUser(re *proto.Sentence) *User {
	return &User{
		Name:            re.Map["name"],
		Password:        re.Map["password"],
		Profile:         re.Map["profile"],
		Comment:         re.Map["comment"],
		Server:          re.Map["server"],
		LimitUptime:     internal.ParseReplyToInt64(re.Map["limit-uptime"]),
		LimitBytesTotal: internal.ParseReplyToInt64(re.Map["limit-bytes-total"]),
		LimitBytesIn:    internal.ParseReplyToInt64(re.Map["limit-bytes-in"]),
		LimitBytesOut:   internal.ParseReplyToInt64(re.Map["limit-bytes-out"]),
		Disabled:        internal.ParseReplyToBool(re.Map["disabled"]),
		Uptime:          re.Map["uptime"],
		BytesIn:         re.Map["bytes-in"],
		BytesOut:        re.Map["bytes-out"],
	}
}

func (c *hotspotClient) buildUserFilterArgs(filter *UserFilter) []string {
	var args []string

	if filter == nil {
		return args
	}

	if filter.Profile != "" {
		args = append(args, "?profile="+filter.Profile)
	}
	// Note: RouterOS ?comment= is an exact match — partial comment searches are
	// handled client-side in GetUsersByComment, so we skip the comment filter here.
	if filter.Disabled != nil {
		disabled := internal.BoolToString(*filter.Disabled)
		args = append(args, "?disabled="+disabled)
	}

	return args
}

func (c *hotspotClient) applyPagination(users []User, offset, limit int) []User {
	start := offset
	if start > len(users) {
		start = len(users)
	}
	end := start + limit
	if end > len(users) {
		end = len(users)
	}
	return users[start:end]
}

func boolPtr(b bool) *bool {
	return &b
}

func stringPtr(s string) *string {
	return &s
}

func int64Ptr(i int64) *int64 {
	return &i
}
