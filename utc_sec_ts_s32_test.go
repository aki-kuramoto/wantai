package wantai

import (
	"testing"
	"time"
)

// Fixed test timestamp: 2024-01-15 12:34:56 UTC as int32.
var testS32Ts = FromTimeS32(time.Date(2024, 1, 15, 12, 34, 56, 0, time.UTC))

// ---------------------------------------------------------------------------
// FromTimeS32 / ToTime
// ---------------------------------------------------------------------------

func TestFromTimeS32(t *testing.T) {
	original := time.Date(2024, 1, 15, 12, 34, 56, 0, time.UTC)
	ts := FromTimeS32(original)
	if int64(ts) != original.Unix() {
		t.Errorf("FromTimeS32() = %d, want %d", int64(ts), original.Unix())
	}
}

func TestFromTimeS32_NonUTC(t *testing.T) {
	tokyo, _ := time.LoadLocation("Asia/Tokyo")
	jst := time.Date(2024, 1, 15, 21, 34, 56, 0, tokyo)
	utc := time.Date(2024, 1, 15, 12, 34, 56, 0, time.UTC)
	if FromTimeS32(jst) != FromTimeS32(utc) {
		t.Errorf("FromTimeS32(JST) != FromTimeS32(UTC) for the same instant")
	}
}

func TestToTime_UtcSecTsS32(t *testing.T) {
	original := time.Date(2024, 1, 15, 12, 34, 56, 0, time.UTC)
	ts := FromTimeS32(original)
	got := ts.ToTime()
	if !got.Equal(original) {
		t.Errorf("ToTime() = %v, want %v", got, original)
	}
	if got.Location() != time.UTC {
		t.Errorf("ToTime() location = %v, want UTC", got.Location())
	}
}

func TestFromTimeS32ToTime_RoundTrip(t *testing.T) {
	original := time.Date(2024, 6, 30, 23, 59, 59, 0, time.UTC)
	got := FromTimeS32(original).ToTime()
	if !got.Equal(original) {
		t.Errorf("round-trip mismatch: got %v, want %v", got, original)
	}
}

// ---------------------------------------------------------------------------
// Epoch and pre-epoch
// ---------------------------------------------------------------------------

func TestUtcSecTsS32_Epoch(t *testing.T) {
	ts := UtcSecTsS32(0)
	got := ts.String()
	want := "1970-01-01T00:00:00Z"
	if got != want {
		t.Errorf("epoch: got %q, want %q", got, want)
	}
}

func TestUtcSecTsS32_Negative(t *testing.T) {
	// -1 → 1969-12-31 23:59:59 UTC
	ts := UtcSecTsS32(-1)
	got := ts.String()
	want := "1969-12-31T23:59:59Z"
	if got != want {
		t.Errorf("negative: got %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// String
// ---------------------------------------------------------------------------

func TestString_UtcSecTsS32(t *testing.T) {
	got := testS32Ts.String()
	want := "2024-01-15T12:34:56Z"
	if got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// Render
// ---------------------------------------------------------------------------

func TestRender_UtcSecTsS32_UTC(t *testing.T) {
	got := testS32Ts.Render("UTC")
	want := "2024-01-15T12:34:56Z"
	if got != want {
		t.Errorf("Render(UTC) = %q, want %q", got, want)
	}
}

func TestRender_UtcSecTsS32_Tokyo(t *testing.T) {
	got := testS32Ts.Render("Asia/Tokyo")
	want := "2024-01-15T21:34:56+09:00"
	if got != want {
		t.Errorf("Render(Asia/Tokyo) = %q, want %q", got, want)
	}
}

func TestRender_UtcSecTsS32_Invalid(t *testing.T) {
	got := testS32Ts.Render("Invalid/Timezone")
	wantUTC := testS32Ts.Render("UTC")
	if got != wantUTC {
		t.Errorf("Render(invalid) = %q, want UTC result %q", got, wantUTC)
	}
}

// ---------------------------------------------------------------------------
// RenderWithFormat
// ---------------------------------------------------------------------------

func TestRenderWithFormat_UtcSecTsS32(t *testing.T) {
	tests := []struct {
		name     string
		timezone string
		format   string
		expected string
	}{
		{"YYYY-MM-DD / UTC", "UTC", "YYYY-MM-DD", "2024-01-15"},
		{"YYYY/MM/DD HH:mm:ss / Tokyo", "Asia/Tokyo", "YYYY/MM/DD HH:mm:ss", "2024/01/15 21:34:56"},
		// Second precision: sub-second tokens output zero.
		{"Milliseconds (zero)", "UTC", "HH:mm:ss.SSS", "12:34:56.000"},
		{"Timezone offset UTC", "UTC", "YYYY-MM-DD HH:mm:ss ZZ", "2024-01-15 12:34:56 +0000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gdf, err := NewGeneralDateFormat(tt.format)
			if err != nil {
				t.Fatalf("NewGeneralDateFormat(%q) error: %v", tt.format, err)
			}
			got := testS32Ts.RenderWithFormat(tt.timezone, *gdf)
			if got != tt.expected {
				t.Errorf("RenderWithFormat(%q, %q) = %q, want %q", tt.timezone, tt.format, got, tt.expected)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// DST handling
// ---------------------------------------------------------------------------

func TestRenderWithFormat_UtcSecTsS32_DST(t *testing.T) {
	tests := []struct {
		name     string
		ts       UtcSecTsS32
		expected string
	}{
		{
			name:     "Winter (EST = UTC-5)",
			ts:       FromTimeS32(time.Date(2024, 1, 15, 12, 34, 56, 0, time.UTC)),
			expected: "2024-01-15 07:34:56 EST -0500",
		},
		{
			name:     "Summer (EDT = UTC-4)",
			ts:       FromTimeS32(time.Date(2024, 7, 15, 12, 34, 56, 0, time.UTC)),
			expected: "2024-07-15 08:34:56 EDT -0400",
		},
	}
	gdf, _ := NewGeneralDateFormat("YYYY-MM-DD HH:mm:ss Z ZZ")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ts.RenderWithFormat("America/New_York", *gdf)
			if got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// int32 max boundary (2038-01-19 03:14:07 UTC)
// ---------------------------------------------------------------------------

func TestUtcSecTsS32_MaxValue(t *testing.T) {
	ts := UtcSecTsS32(0x7FFF_FFFF) // 2147483647
	got := ts.String()
	want := "2038-01-19T03:14:07Z"
	if got != want {
		t.Errorf("max int32: got %q, want %q", got, want)
	}
}

func TestUtcSecTsS32_MinValue(t *testing.T) {
	ts := UtcSecTsS32(-2147483648) // 1901-12-13 20:45:52 UTC
	got := ts.String()
	want := "1901-12-13T20:45:52Z"
	if got != want {
		t.Errorf("min int32: got %q, want %q", got, want)
	}
}
