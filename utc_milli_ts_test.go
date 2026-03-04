package wantai

import (
	"testing"
	"time"
)

// Fixed test timestamp: 2024-01-15 12:34:56.123 UTC (millisecond precision).
var testMilliTs = UtcMilliTs(time.Date(2024, 1, 15, 12, 34, 56, 123_000_000, time.UTC).UnixMilli())

// ---------------------------------------------------------------------------
// FromTimeMillis / ToTime
// ---------------------------------------------------------------------------

func TestFromTimeMillis(t *testing.T) {
	original := time.Date(2024, 1, 15, 12, 34, 56, 123_000_000, time.UTC)
	ts := FromTimeMillis(original)
	if int64(ts) != original.UnixMilli() {
		t.Errorf("FromTimeMillis() = %d, want %d", int64(ts), original.UnixMilli())
	}
}

func TestFromTimeMillis_NonUTC(t *testing.T) {
	tokyo, _ := time.LoadLocation("Asia/Tokyo")
	jst := time.Date(2024, 1, 15, 21, 34, 56, 0, tokyo)
	utc := time.Date(2024, 1, 15, 12, 34, 56, 0, time.UTC)
	if FromTimeMillis(jst) != FromTimeMillis(utc) {
		t.Errorf("FromTimeMillis(JST) != FromTimeMillis(UTC) for the same instant")
	}
}

func TestToTime_UtcMilliTs(t *testing.T) {
	original := time.Date(2024, 1, 15, 12, 34, 56, 123_000_000, time.UTC)
	ts := FromTimeMillis(original)
	got := ts.ToTime()
	if !got.Equal(original) {
		t.Errorf("ToTime() = %v, want %v", got, original)
	}
	if got.Location() != time.UTC {
		t.Errorf("ToTime() location = %v, want UTC", got.Location())
	}
}

func TestFromTimeMillisToTime_RoundTrip(t *testing.T) {
	original := time.Date(2024, 6, 30, 23, 59, 59, 999_000_000, time.UTC)
	got := FromTimeMillis(original).ToTime()
	if !got.Equal(original) {
		t.Errorf("round-trip mismatch: got %v, want %v", got, original)
	}
}

// ---------------------------------------------------------------------------
// String
// ---------------------------------------------------------------------------

func TestString_UtcMilliTs(t *testing.T) {
	ts := UtcMilliTs(time.Date(2024, 1, 15, 12, 34, 56, 0, time.UTC).UnixMilli())
	got := ts.String()
	want := "2024-01-15T12:34:56Z"
	if got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestString_UtcMilliTs_Zero(t *testing.T) {
	got := UtcMilliTs(0).String()
	want := "1970-01-01T00:00:00Z"
	if got != want {
		t.Errorf("UtcMilliTs(0).String() = %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// Render
// ---------------------------------------------------------------------------

func TestRender_UtcMilliTs_UTC(t *testing.T) {
	ts := UtcMilliTs(time.Date(2024, 1, 15, 12, 34, 56, 0, time.UTC).UnixMilli())
	got := ts.Render("UTC")
	want := "2024-01-15T12:34:56Z"
	if got != want {
		t.Errorf("Render(UTC) = %q, want %q", got, want)
	}
}

func TestRender_UtcMilliTs_Tokyo(t *testing.T) {
	ts := UtcMilliTs(time.Date(2024, 1, 15, 12, 34, 56, 0, time.UTC).UnixMilli())
	got := ts.Render("Asia/Tokyo")
	want := "2024-01-15T21:34:56+09:00"
	if got != want {
		t.Errorf("Render(Asia/Tokyo) = %q, want %q", got, want)
	}
}

func TestRender_UtcMilliTs_Invalid(t *testing.T) {
	ts := UtcMilliTs(time.Date(2024, 1, 15, 12, 34, 56, 0, time.UTC).UnixMilli())
	got := ts.Render("Invalid/Timezone")
	wantUTC := ts.Render("UTC")
	if got != wantUTC {
		t.Errorf("Render(invalid) = %q, want UTC result %q", got, wantUTC)
	}
}

// ---------------------------------------------------------------------------
// RenderWithFormat
// ---------------------------------------------------------------------------

func TestRenderWithFormat_UtcMilliTs(t *testing.T) {
	tests := []struct {
		name     string
		timezone string
		format   string
		expected string
	}{
		{"YYYY-MM-DD / UTC", "UTC", "YYYY-MM-DD", "2024-01-15"},
		{"YYYY/MM/DD HH:mm:ss / Tokyo", "Asia/Tokyo", "YYYY/MM/DD HH:mm:ss", "2024/01/15 21:34:56"},
		{"Milliseconds", "UTC", "HH:mm:ss.SSS", "12:34:56.123"},
		// Sub-millisecond digits are 0 for milli precision.
		{"Microseconds (zero)", "UTC", "HH:mm:ss.ffffff", "12:34:56.123000"},
		{"Nanoseconds (zero)", "UTC", "HH:mm:ss.nnnnnnnnn", "12:34:56.123000000"},
		{"Timezone offset UTC", "UTC", "YYYY-MM-DD HH:mm:ss ZZ", "2024-01-15 12:34:56 +0000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gdf, err := NewGeneralDateFormat(tt.format)
			if err != nil {
				t.Fatalf("NewGeneralDateFormat(%q) error: %v", tt.format, err)
			}
			got := testMilliTs.RenderWithFormat(tt.timezone, *gdf)
			if got != tt.expected {
				t.Errorf("RenderWithFormat(%q, %q) = %q, want %q", tt.timezone, tt.format, got, tt.expected)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// DST handling
// ---------------------------------------------------------------------------

func TestRenderWithFormat_UtcMilliTs_DST(t *testing.T) {
	tests := []struct {
		name     string
		ts       UtcMilliTs
		expected string
	}{
		{
			name:     "Winter (EST = UTC-5)",
			ts:       UtcMilliTs(time.Date(2024, 1, 15, 12, 34, 56, 0, time.UTC).UnixMilli()),
			expected: "2024-01-15 07:34:56 EST -0500",
		},
		{
			name:     "Summer (EDT = UTC-4)",
			ts:       UtcMilliTs(time.Date(2024, 7, 15, 12, 34, 56, 0, time.UTC).UnixMilli()),
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
// Boundary: negative timestamp (before epoch)
// ---------------------------------------------------------------------------

func TestRenderWithFormat_UtcMilliTs_PreEpoch(t *testing.T) {
	// 1969-12-31 23:59:59.001 UTC → UnixMilli = -999
	ts := UtcMilliTs(-999)
	gdf, _ := NewGeneralDateFormat("YYYY-MM-DD HH:mm:ss.SSS")
	got := ts.RenderWithFormat("UTC", *gdf)
	want := "1969-12-31 23:59:59.001"
	if got != want {
		t.Errorf("pre-epoch: got %q, want %q", got, want)
	}
}
