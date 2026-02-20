package hotspot

import (
	"strings"
	"testing"
	"time"
)

func TestCalculateExpiryDate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		validity string
		wantDate time.Time
	}{
		{"1d", now.AddDate(0, 0, 1)},
		{"2d", now.AddDate(0, 0, 2)},
		{"30d", now.AddDate(0, 0, 30)},
		{"365d", now.AddDate(0, 0, 365)},
		{"1w", now.AddDate(0, 0, 7)},
		{"2w", now.AddDate(0, 0, 14)},
		{"1y", now.AddDate(1, 0, 0)},
		{"2y", now.AddDate(2, 0, 0)},
		{"2h", now.Add(2 * time.Hour)},
		// "m" follows Go convention: minutes (not months). Use "30d" for ~1 month.
		{"30m", now.Add(30 * time.Minute)},
		{"2h30m", now.Add(2*time.Hour + 30*time.Minute)},
		{"", now}, // empty → today
	}

	for _, tt := range tests {
		t.Run(tt.validity, func(t *testing.T) {
			got := calculateExpiryDate(tt.validity)
			// Parse the returned date string and compare only the date part
			parsed, err := time.Parse("Jan/02/2006", got)
			if err != nil {
				t.Fatalf("calculateExpiryDate(%q) returned unparseable date %q: %v", tt.validity, got, err)
			}

			wantStr := tt.wantDate.Format("Jan/02/2006")
			if got != wantStr {
				t.Errorf("calculateExpiryDate(%q): got %q, want %q", tt.validity, got, wantStr)
			}
			_ = parsed
		})
	}
}

func TestCalculateExpiryDateAlreadyDate(t *testing.T) {
	// If validity is already a formatted date, return as-is
	date := "Jan/20/2026"
	got := calculateExpiryDate(date)
	if got != date {
		t.Errorf("got %q, want %q", got, date)
	}
}

func TestGenerateRandomString(t *testing.T) {
	result := generateRandomString(8, DefaultCharset)
	if len(result) != 8 {
		t.Errorf("expected length 8, got %d", len(result))
	}

	// All chars must be from charset
	for _, c := range result {
		if !strings.ContainsRune(DefaultCharset, c) {
			t.Errorf("char %q not in charset", c)
		}
	}
}

func TestGenerateRandomStringMinLength(t *testing.T) {
	// If length is below MinUsername, clamp to MinUsername
	result := generateRandomString(2, DefaultCharset)
	if len(result) < MinUsername {
		t.Errorf("expected at least %d chars, got %d", MinUsername, len(result))
	}
}

func TestGenerateRandomStringDefaultCharset(t *testing.T) {
	// Empty charset falls back to DefaultCharset
	result := generateRandomString(6, "")
	if len(result) == 0 {
		t.Error("expected non-empty result")
	}
}
