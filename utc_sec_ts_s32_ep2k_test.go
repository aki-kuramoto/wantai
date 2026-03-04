package wantai

import (
	"math"
	"testing"
	"time"
)

// Fixed test timestamp: 2024-01-15 12:34:56 UTC as S32Ep2k.
var testS32Ep2kTs = FromTimeS32Ep2k(time.Date(2024, 1, 15, 12, 34, 56, 0, time.UTC))

// ---------------------------------------------------------------------------
// FromTimeS32Ep2k / ToTime
// ---------------------------------------------------------------------------

func TestFromTimeS32Ep2k(t *testing.T) {
	original := time.Date(2024, 1, 15, 12, 34, 56, 0, time.UTC)
	ts := FromTimeS32Ep2k(original)
	// Value should equal Unix seconds minus the Ep2k offset.
	want := int32(original.Unix() - epoch2kUnixSec)
	if int32(ts) != want {
		t.Errorf("FromTimeS32Ep2k() = %d, want %d", int32(ts), want)
	}
}

func TestFromTimeS32Ep2k_NonUTC(t *testing.T) {
	tokyo, _ := time.LoadLocation("Asia/Tokyo")
	jst := time.Date(2024, 1, 15, 21, 34, 56, 0, tokyo)
	utc := time.Date(2024, 1, 15, 12, 34, 56, 0, time.UTC)
	if FromTimeS32Ep2k(jst) != FromTimeS32Ep2k(utc) {
		t.Errorf("FromTimeS32Ep2k(JST) != FromTimeS32Ep2k(UTC) for the same instant")
	}
}

func TestToTime_UtcSecTsS32Ep2k(t *testing.T) {
	original := time.Date(2024, 1, 15, 12, 34, 56, 0, time.UTC)
	ts := FromTimeS32Ep2k(original)
	got := ts.ToTime()
	if !got.Equal(original) {
		t.Errorf("ToTime() = %v, want %v", got, original)
	}
	if got.Location() != time.UTC {
		t.Errorf("ToTime() location = %v, want UTC", got.Location())
	}
}

func TestFromTimeS32Ep2kToTime_RoundTrip(t *testing.T) {
	original := time.Date(2050, 6, 30, 23, 59, 59, 0, time.UTC)
	got := FromTimeS32Ep2k(original).ToTime()
	if !got.Equal(original) {
		t.Errorf("round-trip mismatch: got %v, want %v", got, original)
	}
}

// ---------------------------------------------------------------------------
// Ep2k epoch origin: UtcSecTsS32Ep2k(0) must equal 2000-01-01T00:00:00Z
// ---------------------------------------------------------------------------

func TestUtcSecTsS32Ep2k_EpochIsYear2000(t *testing.T) {
	ts := UtcSecTsS32Ep2k(0)
	got := ts.String()
	want := "2000-01-01T00:00:00Z"
	if got != want {
		t.Errorf("Ep2k epoch: got %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// Pre-epoch (negative values → before 2000)
// ---------------------------------------------------------------------------

func TestUtcSecTsS32Ep2k_Negative(t *testing.T) {
	// -1 → 1999-12-31T23:59:59Z
	ts := UtcSecTsS32Ep2k(-1)
	got := ts.String()
	want := "1999-12-31T23:59:59Z"
	if got != want {
		t.Errorf("Ep2k(-1): got %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// String
// ---------------------------------------------------------------------------

func TestString_UtcSecTsS32Ep2k(t *testing.T) {
	got := testS32Ep2kTs.String()
	want := "2024-01-15T12:34:56Z"
	if got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// Render
// ---------------------------------------------------------------------------

func TestRender_UtcSecTsS32Ep2k_UTC(t *testing.T) {
	got := testS32Ep2kTs.Render("UTC")
	want := "2024-01-15T12:34:56Z"
	if got != want {
		t.Errorf("Render(UTC) = %q, want %q", got, want)
	}
}

func TestRender_UtcSecTsS32Ep2k_Tokyo(t *testing.T) {
	got := testS32Ep2kTs.Render("Asia/Tokyo")
	want := "2024-01-15T21:34:56+09:00"
	if got != want {
		t.Errorf("Render(Asia/Tokyo) = %q, want %q", got, want)
	}
}

func TestRender_UtcSecTsS32Ep2k_Invalid(t *testing.T) {
	got := testS32Ep2kTs.Render("Invalid/Timezone")
	wantUTC := testS32Ep2kTs.Render("UTC")
	if got != wantUTC {
		t.Errorf("Render(invalid) = %q, want UTC result %q", got, wantUTC)
	}
}

// ---------------------------------------------------------------------------
// RenderWithFormat
// ---------------------------------------------------------------------------

func TestRenderWithFormat_UtcSecTsS32Ep2k(t *testing.T) {
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
			got := testS32Ep2kTs.RenderWithFormat(tt.timezone, *gdf)
			if got != tt.expected {
				t.Errorf("RenderWithFormat(%q, %q) = %q, want %q", tt.timezone, tt.format, got, tt.expected)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// DST handling
// ---------------------------------------------------------------------------

func TestRenderWithFormat_UtcSecTsS32Ep2k_DST(t *testing.T) {
	tests := []struct {
		name     string
		ts       UtcSecTsS32Ep2k
		expected string
	}{
		{
			name:     "Winter (EST = UTC-5)",
			ts:       FromTimeS32Ep2k(time.Date(2024, 1, 15, 12, 34, 56, 0, time.UTC)),
			expected: "2024-01-15 07:34:56 EST -0500",
		},
		{
			name:     "Summer (EDT = UTC-4)",
			ts:       FromTimeS32Ep2k(time.Date(2024, 7, 15, 12, 34, 56, 0, time.UTC)),
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
// int32 boundaries relative to Ep2k epoch
// MaxInt32 → 2068-01-19T03:14:07Z  (= S32 max + 30 years)
// MinInt32 → 1931-12-13T20:45:52Z  (= S32 min + 30 years)
// ---------------------------------------------------------------------------

func TestUtcSecTsS32Ep2k_MaxValue(t *testing.T) {
	ts := UtcSecTsS32Ep2k(math.MaxInt32)
	got := ts.String()
	want := "2068-01-19T03:14:07Z"
	if got != want {
		t.Errorf("max int32 Ep2k: got %q, want %q", got, want)
	}
}

func TestUtcSecTsS32Ep2k_MinValue(t *testing.T) {
	ts := UtcSecTsS32Ep2k(math.MinInt32)
	got := ts.String()
	want := "1931-12-13T20:45:52Z"
	if got != want {
		t.Errorf("min int32 Ep2k: got %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// Verify epoch shift: same wall-clock instant should differ from S32 by 0 seconds
// but the stored integer should differ by epoch2kUnixSec.
// ---------------------------------------------------------------------------

func TestUtcSecTsS32Ep2k_StoredValueOffset(t *testing.T) {
	instant := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	s32 := FromTimeS32(instant)
	ep2k := FromTimeS32Ep2k(instant)

	// Stored integers must differ by epoch2kUnixSec seconds.
	diff := int64(s32) - int64(ep2k)
	if diff != epoch2kUnixSec {
		t.Errorf("stored value difference: got %d, want %d (epoch2kUnixSec)", diff, epoch2kUnixSec)
	}

	// But both render to the same UTC time.
	if s32.String() != ep2k.String() {
		t.Errorf("same instant renders differently:\n S32:    %s\n S32Ep2k: %s", s32.String(), ep2k.String())
	}
}
