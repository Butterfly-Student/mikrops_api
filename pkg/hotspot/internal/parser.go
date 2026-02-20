package internal

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// SaleSeparator is the separator used in sales script names
const SaleSeparator = "-|"

// User mode prefixes
const (
	PrefixVoucher      = "vc-"
	PrefixUserPassword = "up-"
)

// User modes
const (
	ModeVoucher     = "vc"
	ModeUserPassword = "up"
)

// Expiry modes
const (
	ExpiryModeRemove     = "rem"
	ExpiryModeNotify     = "ntf"
	ExpiryModeRemoveCopy = "remc"
	ExpiryModeNotifyCopy = "ntfc"
)

// ProfileData holds profile settings parsed from on-login script
type ProfileData struct {
	ExpiryMode   string
	Price        float64
	Validity     string
	SellingPrice float64
	LockUser     string
}

// SaleData holds sales record information
type SaleData struct {
	Date     string
	Time     string
	Username string
	Price    float64
	Address  string
	Mac      string
	Validity string
}

// ExpiryData holds expiry information from user comment
type ExpiryData struct {
	ExpiryDate time.Time
	UserMode  string
	Prefix    string
}

// ParseOnLoginScript extracts profile settings from RouterOS on-login script
// The script format: :local expmode "rem";:local price "10000";:local validity "1d";...
func ParseOnLoginScript(script string) (*ProfileData, error) {
	profile := &ProfileData{}

	parts := strings.Split(script, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)

		if strings.Contains(part, "expmode") {
			profile.ExpiryMode = extractQuotedValue(part)
		} else if strings.Contains(part, "price") && !strings.Contains(part, "selling") {
			profile.Price = parseFloatFromQuoted(part)
		} else if strings.Contains(part, "validity") {
			profile.Validity = extractQuotedValue(part)
		} else if strings.Contains(part, "selling") {
			profile.SellingPrice = parseFloatFromQuoted(part)
		} else if strings.Contains(part, "lock") {
			profile.LockUser = extractQuotedValue(part)
		}
	}

	return profile, nil
}

// ParseSaleScriptName parses a sales record script name back to SaleData.
// Expected format (7 fixed fields separated by "-|"):
//
//	DATE-|-TIME-|-USERNAME-|-PRICE-|-ADDRESS-|-MAC-|-VALIDITY
//
// Optional fields (ADDRESS, MAC, VALIDITY) may be empty strings.
// Example: "Jan/20/2026-|-16:05:11-|-CAFE-ABCD-|-7000.00-|-192.168.1.1-|-AA:BB-|-1d"
func ParseSaleScriptName(scriptName string) (*SaleData, error) {
	if !strings.Contains(scriptName, SaleSeparator) {
		return nil, fmt.Errorf("invalid sales script format")
	}

	parts := strings.Split(scriptName, SaleSeparator)
	if len(parts) < 4 {
		return nil, fmt.Errorf("invalid sales script format: expected at least 4 fields, got %d", len(parts))
	}

	sale := &SaleData{
		Date:     parts[0],
		Time:     parts[1],
		Username: parts[2],
		Price:    parseFloat(parts[3]),
	}

	// Optional positional fields — always present in new format but may be empty
	if len(parts) > 4 {
		sale.Address = parts[4]
	}
	if len(parts) > 5 {
		sale.Mac = parts[5]
	}
	if len(parts) > 6 {
		sale.Validity = parts[6]
	}

	return sale, nil
}

// ParseUserComment extracts expiry information from user comment
// Format: "DATE vc-" or "DATE up-PREFIX"
func ParseUserComment(comment string) (*ExpiryData, error) {
	if comment == "" {
		return nil, fmt.Errorf("empty comment")
	}

	parts := strings.Fields(comment)
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid comment format")
	}

	// First part should be date
	expiryDate, err := time.Parse("Jan/02/2006", parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	info := &ExpiryData{
		ExpiryDate: expiryDate,
	}

	// Parse user mode and prefix
	if len(parts) > 1 {
		modePrefix := parts[1]
		if strings.HasPrefix(modePrefix, PrefixVoucher) {
			info.UserMode = ModeVoucher
			info.Prefix = strings.TrimPrefix(modePrefix, PrefixVoucher)
		} else if strings.HasPrefix(modePrefix, PrefixUserPassword) {
			info.UserMode = ModeUserPassword
			info.Prefix = strings.TrimPrefix(modePrefix, PrefixUserPassword)
		}
	}

	return info, nil
}

// ParseReplyToInt parses a RouterOS reply field to int
func ParseReplyToInt(value string) int {
	result := 0
	fmt.Sscanf(value, "%d", &result)
	return result
}

// ParseReplyToInt64 parses a RouterOS reply field to int64
func ParseReplyToInt64(value string) int64 {
	result := int64(0)
	fmt.Sscanf(value, "%d", &result)
	return result
}

// ParseReplyToBool parses a RouterOS reply field to bool
func ParseReplyToBool(value string) bool {
	return value == "true" || value == "yes"
}

// BoolToString converts bool to RouterOS yes/no format
func BoolToString(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

// FormatDuration formats seconds to RouterOS duration string (e.g., "1d2h3m")
func FormatDuration(seconds int64) string {
	if seconds == 0 {
		return "0s"
	}

	duration := time.Duration(seconds) * time.Second
	days := int(duration.Hours()) / 24
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd%dh%dm", days, hours, minutes)
	} else if hours > 0 {
		return fmt.Sprintf("%dh%dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

// ParseDuration parses a RouterOS/Mikhmon validity duration string to total seconds.
// Supported formats:
//   - "1d", "2d"     — days (86400 s each)
//   - "1w", "2w"     — weeks (7 days each)
//   - "1h", "2h30m"  — hours and minutes
//   - "30m"          — minutes only
//   - "0s", "unlimited", "" — treated as 0 (unlimited)
//
// Note: month ("1m") and year ("1y") are not converted to seconds here because
// calendar lengths are variable; use calculateExpiryDate in voucher.go for those.
func ParseDuration(dur string) int64 {
	if dur == "" || dur == "0s" || dur == "unlimited" {
		return 0
	}

	lower := strings.ToLower(strings.TrimSpace(dur))

	// Weeks: e.g. "1w"
	if strings.HasSuffix(lower, "w") && !strings.Contains(lower, "d") &&
		!strings.Contains(lower, "h") && !strings.Contains(lower, "m") {
		var weeks int
		fmt.Sscanf(lower, "%dw", &weeks)
		return int64(weeks * 7 * 24 * 60 * 60)
	}

	// RouterOS compound duration: e.g. "1d2h3m", "2h30m", "30m"
	var days, hours, minutes int

	if idx := strings.Index(lower, "d"); idx >= 0 {
		fmt.Sscanf(lower[:idx], "%d", &days)
		lower = lower[idx+1:]
	}
	if idx := strings.Index(lower, "h"); idx >= 0 {
		fmt.Sscanf(lower[:idx], "%d", &hours)
		lower = lower[idx+1:]
	}
	if idx := strings.Index(lower, "m"); idx >= 0 {
		fmt.Sscanf(lower[:idx], "%d", &minutes)
	}

	totalSeconds := days*24*60*60 + hours*60*60 + minutes*60
	return int64(totalSeconds)
}

// FormatBytes formats bytes to human readable string
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// ParseBytes parses human readable bytes string to int64
func ParseBytes(s string) int64 {
	if s == "" || s == "0" || s == "unlimited" {
		return 0
	}

	// Handle formats like: 100M, 1G, 512K, etc.
	multipliers := map[string]int64{
		"B":  1,
		"KB": 1024,
		"K":  1024,
		"MB": 1024 * 1024,
		"M":  1024 * 1024,
		"GB": 1024 * 1024 * 1024,
		"G":  1024 * 1024 * 1024,
		"TB": 1024 * 1024 * 1024 * 1024,
		"T":  1024 * 1024 * 1024 * 1024,
	}

	for suffix, mult := range multipliers {
		if strings.HasSuffix(strings.ToUpper(s), suffix) {
			numStr := strings.TrimSuffix(strings.ToUpper(s), suffix)
			num, _ := strconv.ParseFloat(strings.TrimSpace(numStr), 64)
			return int64(num * float64(mult))
		}
	}

	// Try parsing as plain number
	result, _ := strconv.ParseInt(s, 10, 64)
	return result
}

// Helper functions
func extractQuotedValue(s string) string {
	start := strings.Index(s, `"`)
	end := strings.LastIndex(s, `"`)
	if start >= 0 && end > start {
		return s[start+1 : end]
	}
	return ""
}

func parseFloatFromQuoted(s string) float64 {
	val := extractQuotedValue(s)
	result := 0.0
	fmt.Sscanf(val, "%f", &result)
	return result
}

func parseFloat(s string) float64 {
	result := 0.0
	fmt.Sscanf(s, "%f", &result)
	return result
}
