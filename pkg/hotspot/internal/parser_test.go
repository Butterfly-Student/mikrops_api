package internal

import (
	"testing"
	"time"
)

func TestParseOnLoginScript(t *testing.T) {
	script := `:local expmode "rem";:local price "10000.00";:local validity "1d";:local selling "12000.00";:local lock "yes";`

	data, err := ParseOnLoginScript(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if data.ExpiryMode != "rem" {
		t.Errorf("ExpiryMode: got %q, want %q", data.ExpiryMode, "rem")
	}
	if data.Price != 10000.00 {
		t.Errorf("Price: got %.2f, want %.2f", data.Price, 10000.00)
	}
	if data.Validity != "1d" {
		t.Errorf("Validity: got %q, want %q", data.Validity, "1d")
	}
	if data.SellingPrice != 12000.00 {
		t.Errorf("SellingPrice: got %.2f, want %.2f", data.SellingPrice, 12000.00)
	}
	if data.LockUser != "yes" {
		t.Errorf("LockUser: got %q, want %q", data.LockUser, "yes")
	}
}

func TestBuildAndParseOnLoginScript(t *testing.T) {
	script := BuildOnLoginScript("ntfc", 5000, 7500, "2d", "no")
	data, err := ParseOnLoginScript(script)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if data.ExpiryMode != "ntfc" {
		t.Errorf("ExpiryMode: got %q, want %q", data.ExpiryMode, "ntfc")
	}
	if data.Price != 5000.00 {
		t.Errorf("Price: got %.2f, want 5000.00", data.Price)
	}
	if data.SellingPrice != 7500.00 {
		t.Errorf("SellingPrice: got %.2f, want 7500.00", data.SellingPrice)
	}
	if data.Validity != "2d" {
		t.Errorf("Validity: got %q, want %q", data.Validity, "2d")
	}
	if data.LockUser != "no" {
		t.Errorf("LockUser: got %q, want %q", data.LockUser, "no")
	}
}

func TestBuildAndParseSaleScriptName(t *testing.T) {
	date := "Jan/20/2026"
	saleTime := "16:05:11"
	username := "CAFE-ABCD1234"
	price := 7000.00
	address := "192.168.1.100"
	mac := "AA:BB:CC:DD:EE:FF"
	validity := "1d"

	name := BuildSaleScriptName(date, saleTime, username, price, address, mac, validity)

	data, err := ParseSaleScriptName(name)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if data.Date != date {
		t.Errorf("Date: got %q, want %q", data.Date, date)
	}
	if data.Time != saleTime {
		t.Errorf("Time: got %q, want %q", data.Time, saleTime)
	}
	if data.Username != username {
		t.Errorf("Username: got %q, want %q", data.Username, username)
	}
	if data.Price != price {
		t.Errorf("Price: got %.2f, want %.2f", data.Price, price)
	}
	if data.Address != address {
		t.Errorf("Address: got %q, want %q", data.Address, address)
	}
	if data.Mac != mac {
		t.Errorf("Mac: got %q, want %q", data.Mac, mac)
	}
	if data.Validity != validity {
		t.Errorf("Validity: got %q, want %q", data.Validity, validity)
	}
}

func TestBuildAndParseSaleScriptNameEmptyOptional(t *testing.T) {
	// Empty address and mac — positional fields should still parse correctly
	name := BuildSaleScriptName("Jan/20/2026", "10:00:00", "USER-ABC", 5000, "", "", "2d")
	data, err := ParseSaleScriptName(name)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if data.Username != "USER-ABC" {
		t.Errorf("Username: got %q, want %q", data.Username, "USER-ABC")
	}
	if data.Address != "" {
		t.Errorf("Address should be empty, got %q", data.Address)
	}
	if data.Mac != "" {
		t.Errorf("Mac should be empty, got %q", data.Mac)
	}
	if data.Validity != "2d" {
		t.Errorf("Validity: got %q, want %q", data.Validity, "2d")
	}
}

func TestParseUserComment(t *testing.T) {
	tests := []struct {
		comment    string
		wantMode   string
		wantPrefix string
	}{
		{"Jan/20/2026 vc-CAFE", ModeVoucher, "CAFE"},
		{"Jan/20/2026 vc-", ModeVoucher, ""},
		{"Jan/20/2026 up-SHOP", ModeUserPassword, "SHOP"},
		{"Jan/20/2026 up-", ModeUserPassword, ""},
	}

	for _, tt := range tests {
		t.Run(tt.comment, func(t *testing.T) {
			data, err := ParseUserComment(tt.comment)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if data.UserMode != tt.wantMode {
				t.Errorf("UserMode: got %q, want %q", data.UserMode, tt.wantMode)
			}
			if data.Prefix != tt.wantPrefix {
				t.Errorf("Prefix: got %q, want %q", data.Prefix, tt.wantPrefix)
			}
		})
	}
}

func TestBuildUserComment(t *testing.T) {
	date := "Jan/20/2026"
	comment := BuildUserComment(date, PrefixVoucher, "CAFE")
	if comment != "Jan/20/2026 vc-CAFE" {
		t.Errorf("got %q, want %q", comment, "Jan/20/2026 vc-CAFE")
	}

	comment2 := BuildUserComment(date, PrefixVoucher, "")
	if comment2 != "Jan/20/2026 vc-" {
		t.Errorf("got %q, want %q", comment2, "Jan/20/2026 vc-")
	}
}

func TestExtractPrefix(t *testing.T) {
	tests := []struct {
		username string
		want     string
	}{
		{"CAFE-ABC123", "CAFE"},
		{"SHOP-XYZ-999", "SHOP"},
		{"ABC123", ""},    // no prefix separator
		{"-ABC123", ""},   // starts with separator — invalid, no prefix
	}

	for _, tt := range tests {
		got := ExtractPrefix(tt.username)
		if got != tt.want {
			t.Errorf("ExtractPrefix(%q): got %q, want %q", tt.username, got, tt.want)
		}
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{"1d", 86400},
		{"2d", 172800},
		{"1w", 7 * 86400},
		{"2w", 14 * 86400},
		{"1h", 3600},
		{"30m", 1800},
		{"1d2h3m", 86400 + 7200 + 180},
		{"2h30m", 7200 + 1800},
		{"0s", 0},
		{"unlimited", 0},
		{"", 0},
	}

	for _, tt := range tests {
		got := ParseDuration(tt.input)
		if got != tt.want {
			t.Errorf("ParseDuration(%q): got %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestGetUserMode(t *testing.T) {
	// Voucher: username == password
	if got := GetUserMode("ABC123", "ABC123"); got != PrefixVoucher {
		t.Errorf("got %q, want %q", got, PrefixVoucher)
	}
	// User-password: username != password
	if got := GetUserMode("user1", "pass1"); got != PrefixUserPassword {
		t.Errorf("got %q, want %q", got, PrefixUserPassword)
	}
}

func TestParseReplyToBool(t *testing.T) {
	// RouterOS uses "yes"/"no" and "true"/"false"
	if !ParseReplyToBool("yes") {
		t.Error("expected true for 'yes'")
	}
	if !ParseReplyToBool("true") {
		t.Error("expected true for 'true'")
	}
	if ParseReplyToBool("no") {
		t.Error("expected false for 'no'")
	}
	if ParseReplyToBool("false") {
		t.Error("expected false for 'false'")
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1024 * 1024, "1.0 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
	}

	for _, tt := range tests {
		got := FormatBytes(tt.bytes)
		if got != tt.want {
			t.Errorf("FormatBytes(%d): got %q, want %q", tt.bytes, got, tt.want)
		}
	}
}

// TestCalculateExpiryDateViaBuildUserComment is an integration-style test
// that verifies the validity → expiry date calculation is correct.
func TestBuildExpirySchedulerScriptNotEmpty(t *testing.T) {
	script := BuildExpirySchedulerScript("prepaid-daily")
	if script == "" {
		t.Error("expected non-empty scheduler script")
	}
	if len(script) < 50 {
		t.Errorf("scheduler script suspiciously short: %q", script)
	}
}

func TestBuildOnLoginScript(t *testing.T) {
	script := BuildOnLoginScript("rem", 10000, 12000, "1d", "yes")
	expected := `:local expmode "rem";:local price "10000.00";:local validity "1d";:local selling "12000.00";:local lock "yes";`
	if script != expected {
		t.Errorf("got:\n  %q\nwant:\n  %q", script, expected)
	}
}

func TestBuildSchedulerName(t *testing.T) {
	if got := BuildSchedulerName("prepaid-daily"); got != "monitor-prepaid-daily" {
		t.Errorf("got %q, want %q", got, "monitor-prepaid-daily")
	}
}

func TestParseUserCommentExpiryDate(t *testing.T) {
	comment := "Jan/20/2026 vc-CAFE"
	data, err := ParseUserComment(comment)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := time.Date(2026, time.January, 20, 0, 0, 0, 0, time.UTC)
	if !data.ExpiryDate.Equal(want) {
		t.Errorf("ExpiryDate: got %v, want %v", data.ExpiryDate, want)
	}
}
