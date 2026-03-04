package wantai

import (
	"testing"
	"time"
)

// Fixed test timestamps
var testTimestamp = UtcNanoTs(time.Date(2024, 1, 15, 12, 34, 56, 0, time.UTC).UnixNano())

// testTimestampNano is 2024-01-15 12:34:56.123456789 UTC.
var testTimestampNano = UtcNanoTs(
	time.Date(2024, 1, 15, 12, 34, 56, 123456789, time.UTC).UnixNano(),
)

// ---------------------------------------------------------------------------
// FromTime / ToTime
// ---------------------------------------------------------------------------

func TestFromTime(t *testing.T) {
	original := time.Date(2024, 1, 15, 12, 34, 56, 123456789, time.UTC)
	ts := FromTime(original)
	if int64(ts) != original.UnixNano() {
		t.Errorf("FromTime() = %d, want %d", int64(ts), original.UnixNano())
	}
}

func TestFromTime_NonUTC(t *testing.T) {
	tokyo, _ := time.LoadLocation("Asia/Tokyo")
	jst := time.Date(2024, 1, 15, 21, 34, 56, 0, tokyo)
	utc := time.Date(2024, 1, 15, 12, 34, 56, 0, time.UTC)
	if FromTime(jst) != FromTime(utc) {
		t.Errorf("FromTime(JST) != FromTime(UTC) for the same instant")
	}
}

func TestToTime(t *testing.T) {
	original := time.Date(2024, 1, 15, 12, 34, 56, 0, time.UTC)
	ts := FromTime(original)
	got := ts.ToTime()
	if !got.Equal(original) {
		t.Errorf("ToTime() = %v, want %v", got, original)
	}
	if got.Location() != time.UTC {
		t.Errorf("ToTime() location = %v, want UTC", got.Location())
	}
}

func TestFromTimeToTime_RoundTrip(t *testing.T) {
	original := time.Date(2024, 6, 30, 23, 59, 59, 999999999, time.UTC)
	ts := FromTime(original)
	got := ts.ToTime()
	if !got.Equal(original) {
		t.Errorf("round-trip mismatch: got %v, want %v", got, original)
	}
}

// ---------------------------------------------------------------------------
// String
// ---------------------------------------------------------------------------

func TestString_UtcNanoTs(t *testing.T) {
	got := testTimestamp.String()
	want := "2024-01-15T12:34:56Z"
	if got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestString_UtcNanoTs_Zero(t *testing.T) {
	got := UtcNanoTs(0).String()
	want := "1970-01-01T00:00:00Z"
	if got != want {
		t.Errorf("UtcNanoTs(0).String() = %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// Render
// ---------------------------------------------------------------------------

func TestRender_RFC3339(t *testing.T) {
	got := testTimestamp.Render("UTC")
	want := "2024-01-15T12:34:56Z"
	if got != want {
		t.Errorf("Render(UTC) = %q, want %q", got, want)
	}
}

func TestRender_Zero(t *testing.T) {
	zero := UtcNanoTs(0)
	got := zero.Render("UTC")
	want := "1970-01-01T00:00:00Z"
	if got != want {
		t.Errorf("Render(0) = %q, want %q", got, want)
	}
}

func TestRender_InvalidTimezoneDefaultsToUTC(t *testing.T) {
	got := testTimestamp.Render("Invalid/Timezone")
	wantUTC := testTimestamp.Render("UTC")
	if got != wantUTC {
		t.Errorf("Render with invalid timezone = %q, want UTC result %q", got, wantUTC)
	}
}

// ---------------------------------------------------------------------------
// RenderWithFormat — general patterns
// ---------------------------------------------------------------------------

func TestRenderWithFormat(t *testing.T) {
	tests := []struct {
		name     string
		timezone string
		format   string
		expected string
	}{
		{
			name:     "YYYY-MM-DD / UTC",
			timezone: "UTC",
			format:   "YYYY-MM-DD",
			expected: "2024-01-15",
		},
		{
			name:     "YYYY/MM/DD HH:mm:ss / Tokyo",
			timezone: "Asia/Tokyo",
			format:   "YYYY/MM/DD HH:mm:ss",
			expected: "2024/01/15 21:34:56",
		},
		{
			name:     "ISO8601 with T",
			timezone: "UTC",
			format:   "YYYY-MM-DD'T'HH:mm:ss",
			expected: "2024-01-15T12:34:56",
		},
		{
			name:     "Human-readable",
			timezone: "UTC",
			format:   "DD MMM YYYY",
			expected: "15 Jan 2024",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gdf, err := NewGeneralDateFormat(tt.format)
			if err != nil {
				t.Fatalf("NewGeneralDateFormat(%q) error: %v", tt.format, err)
			}
			got := testTimestamp.RenderWithFormat(tt.timezone, *gdf)
			if got != tt.expected {
				t.Errorf("RenderWithFormat(%q, %q) = %q, want %q", tt.timezone, tt.format, got, tt.expected)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// RenderWithFormat — real patterns with nanosecond timestamp
// ---------------------------------------------------------------------------

func TestRenderWithFormat_RealPatterns(t *testing.T) {
	tests := []struct {
		format   string
		expected string
	}{
		// ISO 8601
		{"YYYY-MM-DD'T'HH:mm:ss", "2024-01-15T12:34:56"},
		{"YYYY-MM-DD'T'HH:mm:ssZZ", "2024-01-15T12:34:56+0000"},
		// US formats
		{"MM/DD/YYYY", "01/15/2024"},
		{"M/D/YY", "1/15/24"},
		// European format
		{"DD.MM.YYYY", "15.01.2024"},
		// Long date
		{"dddd, D MMMM YYYY", "Monday, 15 January 2024"},
		// 12-hour clock
		{"hh:mm:ss A", "12:34:56 PM"},
		{"hh:mm:ss a", "12:34:56 pm"},
		// Apache Common Log format
		{"[DD/MMM/YYYY:HH:mm:ss ZZ]", "[15/Jan/2024:12:34:56 +0000]"},
		// Fractional seconds
		{"HH:mm:ss.SSS", "12:34:56.123"},
		{"HH:mm:ss.ffffff", "12:34:56.123456"},
		{"HH:mm:ss.nnnnnnnnn", "12:34:56.123456789"},
		// Segment tokens
		{"HH:mm:ss.SSSfff", "12:34:56.123456"},
		{"HH:mm:ss.SSSfffnnn", "12:34:56.123456789"},
		{"HH:mm:ss.SSS'***'nnn", "12:34:56.123***789"},
		// Year and time only
		{"YYYY", "2024"},
		{"HH:mm", "12:34"},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			gdf, err := NewGeneralDateFormat(tt.format)
			if err != nil {
				t.Fatalf("NewGeneralDateFormat(%q) error: %v", tt.format, err)
			}
			got := testTimestampNano.RenderWithFormat("UTC", *gdf)
			if got != tt.expected {
				t.Errorf("RenderWithFormat(%q) = %q, want %q", tt.format, got, tt.expected)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// DST timezone (America/New_York)
// ---------------------------------------------------------------------------

func TestRenderWithFormat_DST_NewYork(t *testing.T) {
	tests := []struct {
		name     string
		ts       UtcNanoTs
		format   string
		expected string
	}{
		{
			name:     "Winter (EST = UTC-5)",
			ts:       UtcNanoTs(time.Date(2024, 1, 15, 12, 34, 56, 0, time.UTC).UnixNano()),
			format:   "YYYY-MM-DD HH:mm:ss Z ZZ",
			expected: "2024-01-15 07:34:56 EST -0500",
		},
		{
			name:     "Summer (EDT = UTC-4)",
			ts:       UtcNanoTs(time.Date(2024, 7, 15, 12, 34, 56, 0, time.UTC).UnixNano()),
			format:   "YYYY-MM-DD HH:mm:ss Z ZZ",
			expected: "2024-07-15 08:34:56 EDT -0400",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gdf, err := NewGeneralDateFormat(tt.format)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := tt.ts.RenderWithFormat("America/New_York", *gdf)
			if got != tt.expected {
				t.Errorf("RenderWithFormat(New_York) = %q, want %q", got, tt.expected)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// DST timezone: Southern Hemisphere (Australia/Sydney, Pacific/Auckland)
// ---------------------------------------------------------------------------

func TestRenderWithFormat_DST_Southern(t *testing.T) {
	tests := []struct {
		name     string
		timezone string
		ts       UtcNanoTs
		format   string
		expected string
	}{
		{
			name:     "Sydney winter (AEST = UTC+10)",
			timezone: "Australia/Sydney",
			ts:       UtcNanoTs(time.Date(2024, 7, 15, 12, 34, 56, 0, time.UTC).UnixNano()),
			format:   "YYYY-MM-DD HH:mm:ss Z ZZ",
			expected: "2024-07-15 22:34:56 AEST +1000",
		},
		{
			name:     "Sydney summer (AEDT = UTC+11)",
			timezone: "Australia/Sydney",
			ts:       UtcNanoTs(time.Date(2024, 1, 15, 12, 34, 56, 0, time.UTC).UnixNano()),
			format:   "YYYY-MM-DD HH:mm:ss Z ZZ",
			expected: "2024-01-15 23:34:56 AEDT +1100",
		},
		{
			name:     "Auckland winter (NZST = UTC+12)",
			timezone: "Pacific/Auckland",
			ts:       UtcNanoTs(time.Date(2024, 7, 15, 12, 34, 56, 0, time.UTC).UnixNano()),
			format:   "YYYY-MM-DD HH:mm:ss Z ZZ",
			expected: "2024-07-16 00:34:56 NZST +1200",
		},
		{
			name:     "Auckland summer (NZDT = UTC+13)",
			timezone: "Pacific/Auckland",
			ts:       UtcNanoTs(time.Date(2024, 1, 15, 12, 34, 56, 0, time.UTC).UnixNano()),
			format:   "YYYY-MM-DD HH:mm:ss Z ZZ",
			expected: "2024-01-16 01:34:56 NZDT +1300",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gdf, err := NewGeneralDateFormat(tt.format)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := tt.ts.RenderWithFormat(tt.timezone, *gdf)
			if got != tt.expected {
				t.Errorf("RenderWithFormat(%q) = %q, want %q", tt.timezone, got, tt.expected)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Boundary value tests (int64 limits: ≈ 1677-09-21 and 2262-04-11)
// Uses runNanoBoundaryTests and boundary data from utc_ts_helpers_test.go
// ---------------------------------------------------------------------------

func TestRenderWithFormat_Boundary_1677_UTC(t *testing.T) {
	runNanoBoundaryTests(t, "UTC", boundaryTimestamps1677)
}

func TestRenderWithFormat_Boundary_2262_UTC(t *testing.T) {
	runNanoBoundaryTests(t, "UTC", boundaryTimestamps2262)
}

func TestRenderWithFormat_Boundary_1677_DST(t *testing.T) {
	runNanoBoundaryTests(t, "America/New_York", boundaryTimestamps1677)
}

func TestRenderWithFormat_Boundary_2262_DST(t *testing.T) {
	runNanoBoundaryTests(t, "America/New_York", boundaryTimestamps2262)
}
