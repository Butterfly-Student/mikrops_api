package internal

import (
	"fmt"
	"strings"
)

// BuildOnLoginScript creates a RouterOS on-login script from profile parameters
// This script stores pricing and validity information in the profile itself
func BuildOnLoginScript(expiryMode string, price, sellingPrice float64, validity, lockUser string) string {
	return fmt.Sprintf(
		`:local expmode "%s";:local price "%.2f";:local validity "%s";:local selling "%.2f";:local lock "%s";`,
		expiryMode,
		price,
		validity,
		sellingPrice,
		lockUser,
	)
}

// BuildSaleScriptName creates a RouterOS script name for storing sales records.
// Format: DATE-|-TIME-|-USERNAME-|-PRICE-|-ADDRESS-|-MAC-|-VALIDITY
//
// All 7 fields are always present (optional fields may be empty strings) so that
// ParseSaleScriptName can reliably reconstruct the record by positional index.
// Example: "Jan/20/2026-|-16:05:11-|-CAFE-ABCD-|-7000.00-|-192.168.1.1-|-AA:BB:CC:DD:EE:FF-|-1d"
func BuildSaleScriptName(date, saleTime, username string, price float64, address, mac, validity string) string {
	return fmt.Sprintf("%s%s%s%s%s%s%.2f%s%s%s%s%s%s",
		date, SaleSeparator,
		saleTime, SaleSeparator,
		username, SaleSeparator,
		price, SaleSeparator,
		address, SaleSeparator,
		mac, SaleSeparator,
		validity,
	)
}

// BuildExpirySchedulerScript creates a RouterOS script for monitoring expired users
// in a given profile. The script:
//  1. Iterates every user in the profile
//  2. Checks the comment for a date prefix matching "Mmm/DD/YYYY" (e.g. "Jan/15/2024")
//  3. Compares that date with the current router date
//  4. Reads the profile's expiry mode from its on-login script variable "expmode"
//  5. Removes the user for "rem"/"remc" modes; disables (notifies) for "ntf"/"ntfc"
//
// The on-login variable format expected: `:local expmode "rem";...`
// RouterOS regex note: ^ anchors the match; [A-Z][a-z]{2} matches month abbreviations.
func BuildExpirySchedulerScript(profileName string) string {
	// Use a multi-line string for readability; RouterOS executes it as one statement block.
	return fmt.Sprintf(
		`:local profile "%s";`+
			`:local ponlogin [/ip hotspot user profile get [find where name=$profile] on-login];`+
			// Extract expmode value from ":local expmode "rem";" — pick between the two quotes
			`:local estart ([:find $ponlogin "expmode \""]+9);`+
			`:local eend [:find $ponlogin "\";" $estart];`+
			`:local expmode [:pick $ponlogin $estart $eend];`+
			`/ip hotspot user {`+
			`:foreach i in=[find profile=$profile] do={`+
			`:local cmt [get $i comment];`+
			// Regex matches comment starting with uppercase month abbreviation: "Jan/15/2024 vc-"
			`:if ([:len $cmt] > 0 && $cmt ~ "^[A-Z][a-z][a-z]/[0-9]+/[0-9]+") do={`+
			`:local expdate [:pick $cmt 0 [:find $cmt " "]];`+
			`:local now [/system clock get date];`+
			`:if ($expdate < $now) do={`+
			`:if ($expmode = "rem" || $expmode = "remc") do={ remove $i };`+
			`:if ($expmode = "ntf" || $expmode = "ntfc") do={ set disabled=yes $i };`+
			`}}}}`+
			`}`,
		profileName,
	)
}

// BuildUserComment creates a user comment with expiry date and mode info
// Format: "DATE MODE-PREFIX" or "DATE MODE-"
func BuildUserComment(expiryDate, userMode, prefix string) string {
	return fmt.Sprintf("%s %s%s", expiryDate, userMode, prefix)
}

// BuildSchedulerName creates a scheduler name for profile expiry monitoring
func BuildSchedulerName(profileName string) string {
	return "monitor-" + profileName
}

// GetUserMode determines if a user is voucher mode (vc) or user-password mode (up)
// Voucher: username equals password
// User-Password: username differs from password
func GetUserMode(username, password string) string {
	if username == password {
		return PrefixVoucher
	}
	return PrefixUserPassword
}

// ExtractPrefix extracts the prefix part from a username.
// Username format: "PREFIX-RANDOMPART" → returns "PREFIX".
// If there is no "-" separator the username has no prefix; returns "".
func ExtractPrefix(username string) string {
	idx := strings.Index(username, "-")
	if idx <= 0 {
		return ""
	}
	return username[:idx]
}

// IsValidExpiryMode checks if the given expiry mode is valid
func IsValidExpiryMode(mode string) bool {
	switch mode {
	case ExpiryModeRemove,
		ExpiryModeNotify,
		ExpiryModeRemoveCopy,
		ExpiryModeNotifyCopy:
		return true
	default:
		return false
	}
}

// IsValidUserMode checks if the given user mode is valid
func IsValidUserMode(mode string) bool {
	return mode == ModeVoucher || mode == ModeUserPassword
}
