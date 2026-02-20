package hotspot

const (
	// Character sets for voucher generation
	DefaultCharset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	NumericCharset = "0123456789"
	Alphanumeric   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// Length constraints
	MinUsername = 4
	MaxUsername = 12
	MinPassword = 4
	MaxPassword = 12

	// User modes
	ModeVoucher     = "vc" // username equals password
	ModeUserPassword = "up" // username differs from password

	// Prefix separators
	PrefixSeparator = "-"

	// Sales record separator
	SaleSeparator = "-|"

	// Script owners
	ScriptOwnerSales     = "hotspot-sales"
	ScriptOwnerScheduler = "hotspot-scheduler"

	// Default values
	DefaultServer           = "all"
	DefaultSchedulerInterval = "5m"
	DefaultKeepaliveTimeout = "00:05:00"

	// RouterOS command paths
	PathHotspotUser          = "/ip/hotspot/user"
	PathHotspotUserProfile   = "/ip/hotspot/user/profile"
	PathHotspotActive        = "/ip/hotspot/active"
	PathSystemScheduler      = "/system/scheduler"
	PathSystemScript         = "/system/script"

	// Expiry modes
	ExpiryModeRemove         = "rem"  // Remove user after expiry
	ExpiryModeNotify         = "ntf"  // Notify only after expiry
	ExpiryModeRemoveCopy     = "remc" // Remove and copy after expiry
	ExpiryModeNotifyCopy     = "ntfc" // Notify and copy after expiry
)

// User mode prefixes
const (
	PrefixVoucher      = ModeVoucher + PrefixSeparator
	PrefixUserPassword = ModeUserPassword + PrefixSeparator
)

// Policy permissions for RouterOS scripts/schedulers
const (
	PolicyReadWrite = "read,write,policy,test"
	PolicyReadOnly  = "read"
)
