package wantai

import (
	"testing"
	"time"
)

// Fixed test timestamp: 2024-01-15 12:34:56.123456 UTC (microsecond precision).
var testMicroTs = UtcMicroTs(time.Date(2024, 1, 15, 12, 34, 56, 123_456_000, time.UTC).UnixMicro())

// ---------------------------------------------------------------------------
// FromTimeMicros / ToTime
// ---------------------------------------------------------------------------

func TestFromTimeMicros(t *testing.T) {
	original := time.Date(2024, 1, 15, 12, 34, 56, 123_456_000, time.UTC)
	ts := FromTimeMicros(original)
	if int64(ts) != original.UnixMicro() {
		t.Errorf("FromTimeMicros() = %d, want %d", int64(ts), original.UnixMicro())
	}
}

func TestFromTimeMicros_NonUTC(t *testing.T) {
	tokyo, _ := time.LoadLocation("Asia/Tokyo")
	jst := time.Date(2024, 1, 15, 21, 34, 56, 0, tokyo)
	utc := time.Date(2024, 1, 15, 12, 34, 56, 0, time.UTC)
	if FromTimeMicros(jst) != FromTimeMicros(utc) {
		t.Errorf("FromTimeMicros(JST) != FromTimeMicros(UTC) for the same instant")
	}
}

func TestToTime_UtcMicroTs(t *testing.T) {
	original := time.Date(2024, 1, 15, 12, 34, 56, 123_456_000, time.UTC)
	ts := FromTimeMicros(original)
	got := ts.ToTime()
	if !got.Equal(original) {
		t.Errorf("ToTime() = %v, want %v", got, original)
	}
	if got.Location() != time.UTC {
		t.Errorf("ToTime() location = %v, want UTC", got.Location())
	}
}

func TestFromTimeMicrosToTime_RoundTrip(t *testing.T) {
	original := time.Date(2024, 6, 30, 23, 59, 59, 999_999_000, time.UTC)
	got := FromTimeMicros(original).ToTime()
	if !got.Equal(original) {
		t.Errorf("round-trip mismatch: got %v, want %v", got, original)
	}
}

// ---------------------------------------------------------------------------
// String
// ---------------------------------------------------------------------------

func TestString_UtcMicroTs(t *testing.T) {
	ts := UtcMicroTs(time.Date(2024, 1, 15, 12, 34, 56, 0, time.UTC).UnixMicro())
	got := ts.String()
	want := "2024-01-15T12:34:56Z"
	if got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestString_UtcMicroTs_Zero(t *testing.T) {
	got := UtcMicroTs(0).String()
	want := "1970-01-01T00:00:00Z"
	if got != want {
		t.Errorf("UtcMicroTs(0).String() = %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// Render
// ---------------------------------------------------------------------------

func TestRender_UtcMicroTs_UTC(t *testing.T) {
	ts := UtcMicroTs(time.Date(2024, 1, 15, 12, 34, 56, 0, time.UTC).UnixMicro())
	got := ts.Render("UTC")
	want := "2024-01-15T12:34:56Z"
	if got != want {
		t.Errorf("Render(UTC) = %q, want %q", got, want)
	}
}

func TestRender_UtcMicroTs_Tokyo(t *testing.T) {
	ts := UtcMicroTs(time.Date(2024, 1, 15, 12, 34, 56, 0, time.UTC).UnixMicro())
	got := ts.Render("Asia/Tokyo")
	want := "2024-01-15T21:34:56+09:00"
	if got != want {
		t.Errorf("Render(Asia/Tokyo) = %q, want %q", got, want)
	}
}

func TestRender_UtcMicroTs_Invalid(t *testing.T) {
	ts := UtcMicroTs(time.Date(2024, 1, 15, 12, 34, 56, 0, time.UTC).UnixMicro())
	got := ts.Render("Invalid/Timezone")
	wantUTC := ts.Render("UTC")
	if got != wantUTC {
		t.Errorf("Render(invalid) = %q, want UTC result %q", got, wantUTC)
	}
}

// ---------------------------------------------------------------------------
// RenderWithFormat
// ---------------------------------------------------------------------------

func TestRenderWithFormat_UtcMicroTs(t *testing.T) {
	tests := []struct {
		name     string
		timezone string
		format   string
		expected string
	}{
		{"YYYY-MM-DD / UTC", "UTC", "YYYY-MM-DD", "2024-01-15"},
		{"YYYY/MM/DD HH:mm:ss / Tokyo", "Asia/Tokyo", "YYYY/MM/DD HH:mm:ss", "2024/01/15 21:34:56"},
		{"Milliseconds", "UTC", "HH:mm:ss.SSS", "12:34:56.123"},
		{"Microseconds", "UTC", "HH:mm:ss.ffffff", "12:34:56.123456"},
		// Sub-microsecond nanosecond digits are 0 for micro precision.
		{"Nanoseconds (zero)", "UTC", "HH:mm:ss.nnnnnnnnn", "12:34:56.123456000"},
		{"Timezone offset UTC", "UTC", "YYYY-MM-DD HH:mm:ss ZZ", "2024-01-15 12:34:56 +0000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gdf, err := NewGeneralDateFormat(tt.format)
			if err != nil {
				t.Fatalf("NewGeneralDateFormat(%q) error: %v", tt.format, err)
			}
			got := testMicroTs.RenderWithFormat(tt.timezone, *gdf)
			if got != tt.expected {
				t.Errorf("RenderWithFormat(%q, %q) = %q, want %q", tt.timezone, tt.format, got, tt.expected)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// DST handling
// ---------------------------------------------------------------------------

func TestRenderWithFormat_UtcMicroTs_DST(t *testing.T) {
	tests := []struct {
		name     string
		ts       UtcMicroTs
		expected string
	}{
		{
			name:     "Winter (EST = UTC-5)",
			ts:       UtcMicroTs(time.Date(2024, 1, 15, 12, 34, 56, 0, time.UTC).UnixMicro()),
			expected: "2024-01-15 07:34:56 EST -0500",
		},
		{
			name:     "Summer (EDT = UTC-4)",
			ts:       UtcMicroTs(time.Date(2024, 7, 15, 12, 34, 56, 0, time.UTC).UnixMicro()),
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

func TestRenderWithFormat_UtcMicroTs_PreEpoch(t *testing.T) {
	// 1969-12-31 23:59:59.999999 UTC → UnixMicro = -1
	ts := UtcMicroTs(-1)
	gdf, _ := NewGeneralDateFormat("YYYY-MM-DD HH:mm:ss.ffffff")
	got := ts.RenderWithFormat("UTC", *gdf)
	want := "1969-12-31 23:59:59.999999"
	if got != want {
		t.Errorf("pre-epoch: got %q, want %q", got, want)
	}
}
