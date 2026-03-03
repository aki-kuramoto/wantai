package wantai

import (
	"fmt"
	"math"
	"testing"
	"time"
)

// Fixed test timestamp: 2024-01-15 12:34:56 UTC
var testTimestamp = UtcNanoTs(time.Date(2024, 1, 15, 12, 34, 56, 0, time.UTC).UnixNano())

// testTimestampNano is 2024-01-15 12:34:56.123456789 UTC.
var testTimestampNano = UtcNanoTs(
	time.Date(2024, 1, 15, 12, 34, 56, 123456789, time.UTC).UnixNano(),
)

func TestRender_RFC3339(t *testing.T) {
	got := testTimestamp.Render("UTC")
	want := "2024-01-15T12:34:56Z"
	if got != want {
		t.Errorf("Render(UTC) = %q, want %q", got, want)
	}
}

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

func TestRender_InvalidTimezoneDefaultsToUTC(t *testing.T) {
	got := testTimestamp.Render("Invalid/Timezone")
	wantUTC := testTimestamp.Render("UTC")
	if got != wantUTC {
		t.Errorf("Render with invalid timezone = %q, want UTC result %q", got, wantUTC)
	}
}

func TestClearLocationCache(t *testing.T) {
	// Populate the cache
	_ = testTimestamp.Render("Asia/Tokyo")
	if _, ok := locationCache.Load("Asia/Tokyo"); !ok {
		t.Fatal("Expected Asia/Tokyo to be cached")
	}

	ClearLocationCache()

	if _, ok := locationCache.Load("Asia/Tokyo"); ok {
		t.Error("Expected cache to be empty after ClearLocationCache()")
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
// Real-time rendering end-to-end (testTimestampNano)
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
		// New segment tokens
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
// Summer = October–April (opposite of Northern Hemisphere).
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
// Boundary tests: int64 nanosecond limits (≈ 1677-09-21 and 2262-04-11)
//
// renderExpected uses time.Time field accessors as ground truth and compares
// against our pure-arithmetic renderer.
// ---------------------------------------------------------------------------

const boundaryFormat = "YYYY-MM-DD HH:mm:ss.SSSfffnnn Z ZZ"

// renderExpected computes the reference output using time.Time field accessors
// (NOT time.Format) for the given timestamp and timezone.
func renderExpected(ts UtcNanoTs, loc *time.Location) string {
	sec := int64(ts) / 1_000_000_000
	nano := int64(ts) % 1_000_000_000
	if nano < 0 {
		sec--
		nano += 1_000_000_000
	}
	t := time.Unix(sec, nano).In(loc)

	y, mo, d := t.Date()
	h, mi, s := t.Clock()
	ns := t.Nanosecond()
	tzName, offsetSec := t.Zone()

	ms := ns / 1_000_000
	us := (ns % 1_000_000) / 1_000
	nsub := ns % 1_000

	sign := "+"
	off := offsetSec
	if off < 0 {
		sign = "-"
		off = -off
	}
	return fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d.%03d%03d%03d %s %s%02d%02d",
		y, int(mo), d, h, mi, s, ms, us, nsub, tzName, sign, off/3600, (off%3600)/60)
}

func runBoundaryTests(t *testing.T, timezone string, tss []UtcNanoTs) {
	t.Helper()
	gdf, err := NewGeneralDateFormat(boundaryFormat)
	if err != nil {
		t.Fatalf("NewGeneralDateFormat error: %v", err)
	}
	entry, err := getLocEntryWithCache(timezone)
	if err != nil {
		t.Fatalf("getLocEntryWithCache(%q) error: %v", timezone, err)
	}
	for _, ts := range tss {
		got := ts.RenderWithFormat(timezone, *gdf)
		want := renderExpected(ts, entry.loc)
		if got != want {
			t.Errorf("tz=%q ts=%d:\n got  %q\n want %q", timezone, int64(ts), got, want)
		}
	}
}

// boundaryTimestamps1677 covers the area around int64 nanosecond minimum
// (≈ 1677-09-21 00:12:43 UTC).
var boundaryTimestamps1677 = []UtcNanoTs{
	UtcNanoTs(math.MinInt64),                         // 1677-09-21 minimum
	UtcNanoTs(math.MinInt64 + 1),                     // +1 ns
	UtcNanoTs(math.MinInt64 + 1_000_000_000),         // +1 s
	UtcNanoTs(math.MinInt64 + 86400*1_000_000_000),   // +1 day (~1677-09-22)
	UtcNanoTs(math.MinInt64 + 7*86400*1_000_000_000), // +7 days
}

// boundaryTimestamps2262 covers the area around int64 nanosecond maximum
// (≈ 2262-04-11 23:47:16 UTC).
var boundaryTimestamps2262 = []UtcNanoTs{
	UtcNanoTs(math.MaxInt64),                         // 2262-04-11 maximum
	UtcNanoTs(math.MaxInt64 - 1),                     // -1 ns
	UtcNanoTs(math.MaxInt64 - 1_000_000_000),         // -1 s
	UtcNanoTs(math.MaxInt64 - 86400*1_000_000_000),   // -1 day (~2262-04-10)
	UtcNanoTs(math.MaxInt64 - 7*86400*1_000_000_000), // -7 days
}

func TestRenderWithFormat_Boundary_1677_UTC(t *testing.T) {
	runBoundaryTests(t, "UTC", boundaryTimestamps1677)
}

func TestRenderWithFormat_Boundary_2262_UTC(t *testing.T) {
	runBoundaryTests(t, "UTC", boundaryTimestamps2262)
}

func TestRenderWithFormat_Boundary_1677_DST(t *testing.T) {
	runBoundaryTests(t, "America/New_York", boundaryTimestamps1677)
}

func TestRenderWithFormat_Boundary_2262_DST(t *testing.T) {
	runBoundaryTests(t, "America/New_York", boundaryTimestamps2262)
}
