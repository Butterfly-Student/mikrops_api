package hotspot

import "time"

// Profile represents MikroTik hotspot user profile configuration.
// The profile contains settings for rate limiting, pricing, validity,
// and expiry behavior that are applied to users assigned to this profile.
type Profile struct {
	// Name is the unique profile identifier
	Name string `json:"name"`

	// SharedUsers specifies how many users can share this profile simultaneously
	SharedUsers int `json:"shared_users"`

	// RateLimit defines the bandwidth limit (e.g., "1M/2M" for upload/download)
	RateLimit string `json:"rate_limit"`

	// Validity specifies the time period before user expires (e.g., "1d", "12h")
	Validity string `json:"validity"`

	// Price is the base price for this profile
	Price float64 `json:"price"`

	// SellingPrice is the retail price (can be higher than base Price)
	SellingPrice float64 `json:"selling_price"`

	// ExpiryMode defines behavior after expiry:
	// - "rem": remove user
	// - "ntf": notify only
	// - "remc": remove and copy
	// - "ntfc": notify and copy
	ExpiryMode string `json:"expiry_mode"`

	// LockUser specifies whether to lock user device by MAC address
	LockUser string `json:"lock_user"`

	// KeepaliveTimeout is the idle timeout before disconnecting user
	KeepaliveTimeout string `json:"keepalive_timeout"`

	// OnLoginScript contains RouterOS script executed on user login
	OnLoginScript string `json:"on_login_script,omitempty"`
}

// ProfileUpdate contains updatable profile fields
type ProfileUpdate struct {
	RateLimit        *string
	SharedUsers      *int
	Validity         *string
	Price            *float64
	SellingPrice     *float64
	ExpiryMode       *string
	LockUser         *string
	KeepaliveTimeout *string
}

// User represents a MikroTik hotspot user account.
// Users can be created in two modes: voucher (vc) or user-password (up).
type User struct {
	// Name is the username
	Name string `json:"name"`

	// Password is the authentication password
	Password string `json:"password"`

	// Profile assigns the user to a specific profile
	Profile string `json:"profile"`

	// Comment contains metadata like expiry date and user mode
	// Format: "DATE vc-" or "DATE up-PREFIX"
	Comment string `json:"comment"`

	// LimitUptime is maximum connection time in seconds (0 = unlimited)
	LimitUptime int64 `json:"limit_uptime"`

	// LimitBytesTotal is total data transfer limit in bytes (0 = unlimited)
	LimitBytesTotal int64 `json:"limit_bytes_total"`

	// LimitBytesIn is download limit in bytes
	LimitBytesIn int64 `json:"limit_bytes_in"`

	// LimitBytesOut is upload limit in bytes
	LimitBytesOut int64 `json:"limit_bytes_out"`

	// Disabled indicates if user account is disabled
	Disabled bool `json:"disabled"`

	// Server specifies which hotspot server (or "all")
	Server string `json:"server"`

	// Uptime shows current session duration (read-only)
	Uptime string `json:"uptime,omitempty"`

	// BytesIn shows bytes downloaded (read-only)
	BytesIn string `json:"bytes_in,omitempty"`

	// BytesOut shows bytes uploaded (read-only)
	BytesOut string `json:"bytes_out,omitempty"`
}

// UserUpdate contains updatable user fields
type UserUpdate struct {
	Profile         *string
	Disabled        *bool
	Comment         *string
	LimitUptime     *int64
	LimitBytesTotal *int64
	LimitBytesIn    *int64
	LimitBytesOut   *int64
}

// Session represents an active hotspot user session.
// Sessions are created when users authenticate and removed on logout/disconnect.
type Session struct {
	// Name is the authenticated username
	Name string `json:"name"`

	// Address is the client's IP address
	Address string `json:"address"`

	// MacAddress is the client's MAC address
	MacAddress string `json:"mac_address"`

	// Uptime is the duration of current session
	Uptime string `json:"uptime"`

	// SessionTimeLeft shows remaining time before forced disconnect
	SessionTimeLeft string `json:"session_time_left"`

	// BytesIn shows bytes downloaded in this session
	BytesIn string `json:"bytes_in"`

	// BytesOut shows bytes uploaded in this session
	BytesOut string `json:"bytes_out"`

	// LoginBy indicates authentication method (cookie, httpchap, https)
	LoginBy string `json:"login_by"`
}

// SessionStats contains aggregated statistics for all active sessions
type SessionStats struct {
	// TotalUsers is the total number of user records
	TotalUsers int `json:"total_users"`

	// ActiveUsers is the number of currently active sessions
	ActiveUsers int `json:"active_users"`

	// TotalBytesIn is formatted total download across all sessions
	TotalBytesIn string `json:"total_bytes_in"`

	// TotalBytesOut is formatted total upload across all sessions
	TotalBytesOut string `json:"total_bytes_out"`
}

// Sale represents a sales transaction record stored as RouterOS script.
// This follows Mikhmon v3's pattern of storing sales in script names.
type Sale struct {
	// Date when sale occurred (format: "Jan/02/2006")
	Date string `json:"date"`

	// Time when sale occurred (format: "15:04:05")
	Time string `json:"time"`

	// Username that was sold
	Username string `json:"username"`

	// Price charged for the voucher
	Price float64 `json:"price"`

	// Address is client IP (optional)
	Address string `json:"address,omitempty"`

	// Mac is client MAC address (optional)
	Mac string `json:"mac,omitempty"`

	// Validity period of the voucher
	Validity string `json:"validity,omitempty"`

	// ScriptID is the RouterOS script ID (read-only)
	ScriptID string `json:"script_id,omitempty"`
}

// Scheduler represents a RouterOS scheduler for automated tasks.
// Schedulers are typically used for expiry monitoring.
type Scheduler struct {
	// Name is the scheduler name (usually "monitor-{profile}")
	Name string `json:"name"`

	// Interval is the execution interval (e.g., "5m")
	Interval string `json:"interval"`

	// StartTime is when the scheduler starts (e.g., "startup")
	StartTime string `json:"start_time"`

	// Policy defines RouterOS policy permissions
	Policy string `json:"policy"`

	// OnEvent contains the script to execute
	OnEvent string `json:"on_event"`

	// Enabled indicates if scheduler is active
	Enabled bool `json:"enabled"`
}

// VoucherGenerator contains configuration for batch voucher generation
type VoucherGenerator struct {
	// Profile assigns vouchers to this profile
	Profile string `json:"profile"`

	// Prefix is prepended to generated usernames
	Prefix string `json:"prefix"`

	// Charset defines characters for random generation
	Charset string `json:"charset"`

	// LengthUsername is the length of random username part
	LengthUsername int `json:"length_username"`

	// LengthPassword is the length of password
	LengthPassword int `json:"length_password"`

	// Quantity is the number of vouchers to generate
	Quantity int `json:"quantity"`

	// TimeLimit is session duration in seconds (0 = unlimited)
	TimeLimit int64 `json:"time_limit"`

	// DataLimit is data transfer limit in bytes (0 = unlimited)
	DataLimit int64 `json:"data_limit"`

	// Validity specifies how long vouchers are valid (e.g., "1d", "1w", "1m")
	Validity string `json:"validity"`

	// Mode specifies generation mode (vc or up)
	Mode string `json:"mode"` // "vc" or "up"
}

// VoucherResult contains the results of voucher generation
type VoucherResult struct {
	// Success is the count of successfully created vouchers
	Success int `json:"success"`

	// Failed is the count of failed voucher creations
	Failed int `json:"failed"`

	// Vouchers contains the successfully created vouchers
	Vouchers []User `json:"vouchers"`

	// Errors contains error messages for failed creations
	Errors []string `json:"errors,omitempty"`
}

// UserFilter defines filtering options for user queries
type UserFilter struct {
	// Profile filters by specific profile (RouterOS server-side filter)
	Profile string `json:"profile,omitempty"`

	// Disabled filters by enabled/disabled status (RouterOS server-side filter)
	Disabled *bool `json:"disabled,omitempty"`

	// Limit specifies maximum results to return (client-side pagination)
	Limit int `json:"limit,omitempty"`

	// Offset specifies number of results to skip (client-side pagination)
	Offset int `json:"offset,omitempty"`
}

// SaleFilter defines filtering options for sales queries
type SaleFilter struct {
	// StartDate is the beginning of date range (format: "Jan/02/2006")
	StartDate string `json:"start_date,omitempty"`

	// EndDate is the end of date range (format: "Jan/02/2006")
	EndDate string `json:"end_date,omitempty"`

	// Prefix filters by username prefix
	Prefix string `json:"prefix,omitempty"`

	// Limit specifies maximum results to return
	Limit int `json:"limit,omitempty"`

	// Offset specifies number of results to skip
	Offset int `json:"offset,omitempty"`
}

// Config holds client configuration
type Config struct {
	// RouterID identifies which router this client connects to
	RouterID uint `json:"router_id"`

	// HotspotName is the hotspot server name
	HotspotName string `json:"hotspot_name,omitempty"`

	// Currency for price formatting
	Currency string `json:"currency,omitempty"`

	// Debug enables verbose logging
	Debug bool `json:"debug,omitempty"`

	// DefaultProfile is used when no profile specified
	DefaultProfile string `json:"default_profile,omitempty"`

	// DefaultValidity is used when no validity specified
	DefaultValidity string `json:"default_validity,omitempty"`
}

// ExpiryInfo contains parsed expiry information from user comment
type ExpiryInfo struct {
	// ExpiryDate is when the user expires
	ExpiryDate time.Time `json:"expiry_date"`

	// UserMode is either "vc" (voucher) or "up" (user-password)
	UserMode string `json:"user_mode"`

	// Prefix is the username prefix
	Prefix string `json:"prefix,omitempty"`
}
