package wantai

import (
	"testing"
	"time"
)

// Fixed test timestamp: 2024-01-15 12:34:56 UTC
var testTimestamp = UtcNanoTs(time.Date(2024, 1, 15, 12, 34, 56, 0, time.UTC).UnixNano())

func TestRender_RFC3339(t *testing.T) {
	got := testTimestamp.Render("UTC")
	want := "2024-01-15T12:34:56Z"
	if got != want {
		t.Errorf("Render(UTC) = %q, want %q", got, want)
	}
}

func TestRenderWithGoLayout(t *testing.T) {
	tests := []struct {
		name     string
		timezone string
		layout   string
		expected string
	}{
		{
			name:     "UTC / date only",
			timezone: "UTC",
			layout:   "2006-01-02",
			expected: "2024-01-15",
		},
		{
			name:     "UTC / time only",
			timezone: "UTC",
			layout:   "15:04:05",
			expected: "12:34:56",
		},
		{
			name:     "Asia/Tokyo / datetime",
			timezone: "Asia/Tokyo",
			layout:   "2006-01-02 15:04:05",
			expected: "2024-01-15 21:34:56", // UTC+9
		},
		{
			name:     "America/New_York / datetime",
			timezone: "America/New_York",
			layout:   "2006-01-02 15:04:05",
			expected: "2024-01-15 07:34:56", // UTC-5
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := testTimestamp.RenderWithGoLayout(tt.timezone, tt.layout)
			if got != tt.expected {
				t.Errorf("RenderWithGoLayout(%q, %q) = %q, want %q", tt.timezone, tt.layout, got, tt.expected)
			}
		})
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
	// An invalid timezone name should fall back to UTC
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
	// A zero-value timestamp (Unix epoch) should render correctly
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

	// Should match at nanosecond precision
	if int64(ts) != original.UnixNano() {
		t.Errorf("FromTime() = %d, want %d", int64(ts), original.UnixNano())
	}
}

func TestFromTime_NonUTC(t *testing.T) {
	// A time.Time in a non-UTC timezone should produce the same timestamp as its UTC equivalent
	tokyo, _ := time.LoadLocation("Asia/Tokyo")
	// 2024-01-15 21:34:56 JST (= 2024-01-15 12:34:56 UTC)
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

	// Values should be equal
	if !got.Equal(original) {
		t.Errorf("ToTime() = %v, want %v", got, original)
	}
	// Location should be UTC
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
// Category 4: Real-time rendering (end-to-end)
// Uses a fixed nanosecond-precision timestamp to validate full conversion
// through NewGeneralDateFormat → RenderWithFormat.
// ---------------------------------------------------------------------------

// testTimestampNano is 2024-01-15 12:34:56.123456789 UTC.
var testTimestampNano = UtcNanoTs(
	time.Date(2024, 1, 15, 12, 34, 56, 123456789, time.UTC).UnixNano(),
)

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
		// Fractional seconds (uses testTimestampNano for sub-second values)
		{"HH:mm:ss.SSS", "12:34:56.123"},
		{"HH:mm:ss.ffffff", "12:34:56.123456"},
		{"HH:mm:ss.nnnnnn", "12:34:56.123456789"},
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
