package wantai

import (
	"testing"
	"time"
)

// Fixed test timestamp: 2024-01-15 12:34:56 UTC as U32Ep2k.
var testU32Ep2kTs = FromTimeU32Ep2k(time.Date(2024, 1, 15, 12, 34, 56, 0, time.UTC))

// ---------------------------------------------------------------------------
// FromTimeU32Ep2k / ToTime
// ---------------------------------------------------------------------------

func TestFromTimeU32Ep2k(t *testing.T) {
	original := time.Date(2024, 1, 15, 12, 34, 56, 0, time.UTC)
	ts := FromTimeU32Ep2k(original)
	want := uint32(original.Unix() - epoch2kUnixSec)
	if uint32(ts) != want {
		t.Errorf("FromTimeU32Ep2k() = %d, want %d", uint32(ts), want)
	}
}

func TestFromTimeU32Ep2k_NonUTC(t *testing.T) {
	tokyo, _ := time.LoadLocation("Asia/Tokyo")
	jst := time.Date(2024, 1, 15, 21, 34, 56, 0, tokyo)
	utc := time.Date(2024, 1, 15, 12, 34, 56, 0, time.UTC)
	if FromTimeU32Ep2k(jst) != FromTimeU32Ep2k(utc) {
		t.Errorf("FromTimeU32Ep2k(JST) != FromTimeU32Ep2k(UTC) for the same instant")
	}
}

func TestToTime_UtcSecTsU32Ep2k(t *testing.T) {
	original := time.Date(2024, 1, 15, 12, 34, 56, 0, time.UTC)
	ts := FromTimeU32Ep2k(original)
	got := ts.ToTime()
	if !got.Equal(original) {
		t.Errorf("ToTime() = %v, want %v", got, original)
	}
	if got.Location() != time.UTC {
		t.Errorf("ToTime() location = %v, want UTC", got.Location())
	}
}

func TestFromTimeU32Ep2kToTime_RoundTrip(t *testing.T) {
	original := time.Date(2100, 6, 30, 23, 59, 59, 0, time.UTC)
	got := FromTimeU32Ep2k(original).ToTime()
	if !got.Equal(original) {
		t.Errorf("round-trip mismatch: got %v, want %v", got, original)
	}
}

// ---------------------------------------------------------------------------
// Ep2k epoch origin: UtcSecTsU32Ep2k(0) must equal 2000-01-01T00:00:00Z
// ---------------------------------------------------------------------------

func TestUtcSecTsU32Ep2k_EpochIsYear2000(t *testing.T) {
	ts := UtcSecTsU32Ep2k(0)
	got := ts.String()
	want := "2000-01-01T00:00:00Z"
	if got != want {
		t.Errorf("Ep2k epoch: got %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// String
// ---------------------------------------------------------------------------

func TestString_UtcSecTsU32Ep2k(t *testing.T) {
	got := testU32Ep2kTs.String()
	want := "2024-01-15T12:34:56Z"
	if got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// Render
// ---------------------------------------------------------------------------

func TestRender_UtcSecTsU32Ep2k_UTC(t *testing.T) {
	got := testU32Ep2kTs.Render("UTC")
	want := "2024-01-15T12:34:56Z"
	if got != want {
		t.Errorf("Render(UTC) = %q, want %q", got, want)
	}
}

func TestRender_UtcSecTsU32Ep2k_Tokyo(t *testing.T) {
	got := testU32Ep2kTs.Render("Asia/Tokyo")
	want := "2024-01-15T21:34:56+09:00"
	if got != want {
		t.Errorf("Render(Asia/Tokyo) = %q, want %q", got, want)
	}
}

func TestRender_UtcSecTsU32Ep2k_Invalid(t *testing.T) {
	got := testU32Ep2kTs.Render("Invalid/Timezone")
	wantUTC := testU32Ep2kTs.Render("UTC")
	if got != wantUTC {
		t.Errorf("Render(invalid) = %q, want UTC result %q", got, wantUTC)
	}
}

// ---------------------------------------------------------------------------
// RenderWithFormat
// ---------------------------------------------------------------------------

func TestRenderWithFormat_UtcSecTsU32Ep2k(t *testing.T) {
	tests := []struct {
		name     string
		timezone string
		format   string
		expected string
	}{
		{"YYYY-MM-DD / UTC", "UTC", "YYYY-MM-DD", "2024-01-15"},
		{"YYYY/MM/DD HH:mm:ss / Tokyo", "Asia/Tokyo", "YYYY/MM/DD HH:mm:ss", "2024/01/15 21:34:56"},
		{"Milliseconds (zero)", "UTC", "HH:mm:ss.SSS", "12:34:56.000"},
		{"Timezone offset UTC", "UTC", "YYYY-MM-DD HH:mm:ss ZZ", "2024-01-15 12:34:56 +0000"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gdf, err := NewGeneralDateFormat(tt.format)
			if err != nil {
				t.Fatalf("NewGeneralDateFormat(%q) error: %v", tt.format, err)
			}
			got := testU32Ep2kTs.RenderWithFormat(tt.timezone, *gdf)
			if got != tt.expected {
				t.Errorf("RenderWithFormat(%q, %q) = %q, want %q", tt.timezone, tt.format, got, tt.expected)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// DST handling
// ---------------------------------------------------------------------------

func TestRenderWithFormat_UtcSecTsU32Ep2k_DST(t *testing.T) {
	tests := []struct {
		name     string
		ts       UtcSecTsU32Ep2k
		expected string
	}{
		{
			name:     "Winter (EST = UTC-5)",
			ts:       FromTimeU32Ep2k(time.Date(2024, 1, 15, 12, 34, 56, 0, time.UTC)),
			expected: "2024-01-15 07:34:56 EST -0500",
		},
		{
			name:     "Summer (EDT = UTC-4)",
			ts:       FromTimeU32Ep2k(time.Date(2024, 7, 15, 12, 34, 56, 0, time.UTC)),
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
// uint32 boundaries relative to Ep2k epoch
// MaxUint32 → 2136-02-07T06:28:15Z  (= U32 max + 30 years)
// ---------------------------------------------------------------------------

func TestUtcSecTsU32Ep2k_Beyond2106(t *testing.T) {
	// UtcSecTsU32 max is 2106-02-07; Ep2k pushes this to 2136-02-07.
	ts := UtcSecTsU32Ep2k(0xFFFF_FFFF)
	got := ts.String()
	want := "2136-02-07T06:28:15Z"
	if got != want {
		t.Errorf("max uint32 Ep2k: got %q, want %q", got, want)
	}
}

func TestUtcSecTsU32Ep2k_Year2100(t *testing.T) {
	// 2100-01-01 is representable by both U32 and U32Ep2k; values differ by offset.
	ts := FromTimeU32Ep2k(time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC))
	got := ts.String()
	want := "2100-01-01T00:00:00Z"
	if got != want {
		t.Errorf("year 2100: got %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// Verify epoch shift: same wall-clock instant stores different integers.
// ---------------------------------------------------------------------------

func TestUtcSecTsU32Ep2k_StoredValueOffset(t *testing.T) {
	instant := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	u32 := FromTimeU32(instant)
	ep2k := FromTimeU32Ep2k(instant)

	// Stored integers must differ by epoch2kUnixSec.
	diff := int64(u32) - int64(ep2k)
	if diff != epoch2kUnixSec {
		t.Errorf("stored value difference: got %d, want %d (epoch2kUnixSec)", diff, epoch2kUnixSec)
	}

	// But both render to the same UTC time.
	if u32.String() != ep2k.String() {
		t.Errorf("same instant renders differently:\n U32:     %s\n U32Ep2k: %s", u32.String(), ep2k.String())
	}
}
