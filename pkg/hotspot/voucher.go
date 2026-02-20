package hotspot

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"go-template/pkg/hotspot/internal"
)

// GenerateVouchers generates batch of voucher mode users (username = password)
func (c *hotspotClient) GenerateVouchers(ctx context.Context, gen *VoucherGenerator) (*VoucherResult, error) {
	// Validate
	if gen.Profile == "" {
		return nil, NewError("generate vouchers", fmt.Errorf("profile is required"))
	}
	if gen.Quantity <= 0 {
		return nil, ErrInvalidQuantity
	}

	// Set defaults
	if gen.Charset == "" {
		gen.Charset = DefaultCharset
	}
	if gen.LengthUsername < MinUsername {
		gen.LengthUsername = 8
	}
	if gen.LengthPassword < MinPassword {
		gen.LengthPassword = 8
	}

	result := &VoucherResult{
		Vouchers: make([]User, 0, gen.Quantity),
		Errors:   make([]string, 0),
	}

	for i := 0; i < gen.Quantity; i++ {
		// Generate random part
		randomPart := generateRandomString(gen.LengthUsername, gen.Charset)

		// Build username with prefix (skip separator if prefix is empty)
		var username string
		if gen.Prefix != "" {
			username = gen.Prefix + PrefixSeparator + randomPart
		} else {
			username = randomPart
		}

		// Password equals username for voucher mode
		password := username

		// Build comment with expiry date and mode
		expiryDate := calculateExpiryDate(gen.Validity)
		comment := internal.BuildUserComment(expiryDate, PrefixVoucher, gen.Prefix)

		user := User{
			Name:            username,
			Password:        password,
			Profile:         gen.Profile,
			Comment:         comment,
			LimitUptime:     gen.TimeLimit,
			LimitBytesTotal: gen.DataLimit,
			Disabled:        false,
			Server:          DefaultServer,
		}

		// Create user on RouterOS
		err := c.CreateUser(ctx, &user)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors,
				fmt.Sprintf("%s: %v", username, err))
		} else {
			result.Success++
			result.Vouchers = append(result.Vouchers, user)
		}
	}

	return result, nil
}

// GenerateUserPasswordMode generates batch of user-password mode users (username != password)
func (c *hotspotClient) GenerateUserPasswordMode(ctx context.Context, gen *VoucherGenerator) (*VoucherResult, error) {
	// Validate
	if gen.Profile == "" {
		return nil, NewError("generate users", fmt.Errorf("profile is required"))
	}
	if gen.Quantity <= 0 {
		return nil, ErrInvalidQuantity
	}

	// Set defaults
	if gen.Charset == "" {
		gen.Charset = DefaultCharset
	}
	if gen.LengthUsername < MinUsername {
		gen.LengthUsername = 8
	}
	if gen.LengthPassword < MinPassword {
		gen.LengthPassword = 8
	}

	result := &VoucherResult{
		Vouchers: make([]User, 0, gen.Quantity),
		Errors:   make([]string, 0),
	}

	for i := 0; i < gen.Quantity; i++ {
		randomPart := generateRandomString(gen.LengthUsername, gen.Charset)

		// Build username with prefix (skip separator if prefix is empty)
		var username string
		if gen.Prefix != "" {
			username = gen.Prefix + PrefixSeparator + randomPart
		} else {
			username = randomPart
		}
		password := generateRandomString(gen.LengthPassword, gen.Charset)

		// Build comment with expiry date and mode
		expiryDate := calculateExpiryDate(gen.Validity)
		comment := internal.BuildUserComment(expiryDate, PrefixUserPassword, gen.Prefix)

		user := User{
			Name:            username,
			Password:        password,
			Profile:         gen.Profile,
			Comment:         comment,
			LimitUptime:     gen.TimeLimit,
			LimitBytesTotal: gen.DataLimit,
			Disabled:        false,
			Server:          DefaultServer,
		}

		err := c.CreateUser(ctx, &user)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors,
				fmt.Sprintf("%s: %v", username, err))
		} else {
			result.Success++
			result.Vouchers = append(result.Vouchers, user)
		}
	}

	return result, nil
}

// generateRandomString generates a random string from the given charset.
// Uses the global rand source (auto-seeded since Go 1.20).
func generateRandomString(length int, charset string) string {
	if charset == "" {
		charset = DefaultCharset
	}
	if length < MinUsername {
		length = MinUsername
	}

	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

// calculateExpiryDate calculates the expiry date string from a validity period.
// Supported formats (processed in order):
//
//   - "Jan/02/2006"       — already a formatted date, returned as-is
//   - "Nd" / "Nd"         — N days   (e.g. "1d", "30d")
//   - "Nw"                — N weeks  (e.g. "1w", "2w")
//   - "Ny"                — N years  (e.g. "1y")
//   - Go duration string  — parsed by time.ParseDuration (e.g. "2h", "30m", "2h30m")
//
// Note on "m": to avoid ambiguity with Go's "m = minutes", months are NOT a
// supported single-letter suffix. Use "30d" for ~1 month, or "365d" for ~1 year.
// For calendar-exact months, pass a pre-calculated date string ("Jan/02/2006").
func calculateExpiryDate(validity string) string {
	if validity == "" {
		return time.Now().Format("Jan/02/2006")
	}

	// Already a formatted date
	if _, err := time.Parse("Jan/02/2006", validity); err == nil {
		return validity
	}

	now := time.Now()
	lower := strings.ToLower(strings.TrimSpace(validity))

	// Days: "1d", "30d", "365d" — must have only digits before "d"
	if strings.HasSuffix(lower, "d") && !strings.ContainsAny(lower[:len(lower)-1], "wyhms") {
		var days int
		fmt.Sscanf(lower, "%dd", &days)
		if days > 0 {
			return now.AddDate(0, 0, days).Format("Jan/02/2006")
		}
	}

	// Weeks: "1w", "2w"
	if strings.HasSuffix(lower, "w") {
		var weeks int
		fmt.Sscanf(lower, "%dw", &weeks)
		if weeks > 0 {
			return now.AddDate(0, 0, weeks*7).Format("Jan/02/2006")
		}
	}

	// Years: "1y", "2y"
	if strings.HasSuffix(lower, "y") {
		var years int
		fmt.Sscanf(lower, "%dy", &years)
		if years > 0 {
			return now.AddDate(years, 0, 0).Format("Jan/02/2006")
		}
	}

	// Go standard duration: "2h", "30m", "2h30m", "3600s"
	if d, err := time.ParseDuration(lower); err == nil && d > 0 {
		return now.Add(d).Format("Jan/02/2006")
	}

	// Unparseable — default to today
	return now.Format("Jan/02/2006")
}
