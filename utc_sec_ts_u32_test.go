package wantai

import (
	"testing"
	"time"
)

// Fixed test timestamp: 2024-01-15 12:34:56 UTC as uint32.
var testU32Ts = FromTimeU32(time.Date(2024, 1, 15, 12, 34, 56, 0, time.UTC))

// ---------------------------------------------------------------------------
// FromTimeU32 / ToTime
// ---------------------------------------------------------------------------

func TestFromTimeU32(t *testing.T) {
	original := time.Date(2024, 1, 15, 12, 34, 56, 0, time.UTC)
	ts := FromTimeU32(original)
	if uint64(ts) != uint64(original.Unix()) {
		t.Errorf("FromTimeU32() = %d, want %d", uint64(ts), original.Unix())
	}
}

func TestFromTimeU32_NonUTC(t *testing.T) {
	tokyo, _ := time.LoadLocation("Asia/Tokyo")
	jst := time.Date(2024, 1, 15, 21, 34, 56, 0, tokyo)
	utc := time.Date(2024, 1, 15, 12, 34, 56, 0, time.UTC)
	if FromTimeU32(jst) != FromTimeU32(utc) {
		t.Errorf("FromTimeU32(JST) != FromTimeU32(UTC) for the same instant")
	}
}

func TestToTime_UtcSecTsU32(t *testing.T) {
	original := time.Date(2024, 1, 15, 12, 34, 56, 0, time.UTC)
	ts := FromTimeU32(original)
	got := ts.ToTime()
	if !got.Equal(original) {
		t.Errorf("ToTime() = %v, want %v", got, original)
	}
	if got.Location() != time.UTC {
		t.Errorf("ToTime() location = %v, want UTC", got.Location())
	}
}

func TestFromTimeU32ToTime_RoundTrip(t *testing.T) {
	original := time.Date(2024, 6, 30, 23, 59, 59, 0, time.UTC)
	got := FromTimeU32(original).ToTime()
	if !got.Equal(original) {
		t.Errorf("round-trip mismatch: got %v, want %v", got, original)
	}
}

// ---------------------------------------------------------------------------
// Epoch
// ---------------------------------------------------------------------------

func TestUtcSecTsU32_Epoch(t *testing.T) {
	ts := UtcSecTsU32(0)
	got := ts.String()
	want := "1970-01-01T00:00:00Z"
	if got != want {
		t.Errorf("epoch: got %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// String
// ---------------------------------------------------------------------------

func TestString_UtcSecTsU32(t *testing.T) {
	got := testU32Ts.String()
	want := "2024-01-15T12:34:56Z"
	if got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

// ---------------------------------------------------------------------------
// Render
// ---------------------------------------------------------------------------

func TestRender_UtcSecTsU32_UTC(t *testing.T) {
	got := testU32Ts.Render("UTC")
	want := "2024-01-15T12:34:56Z"
	if got != want {
		t.Errorf("Render(UTC) = %q, want %q", got, want)
	}
}

func TestRender_UtcSecTsU32_Tokyo(t *testing.T) {
	got := testU32Ts.Render("Asia/Tokyo")
	want := "2024-01-15T21:34:56+09:00"
	if got != want {
		t.Errorf("Render(Asia/Tokyo) = %q, want %q", got, want)
	}
}

func TestRender_UtcSecTsU32_Invalid(t *testing.T) {
	got := testU32Ts.Render("Invalid/Timezone")
	wantUTC := testU32Ts.Render("UTC")
	if got != wantUTC {
		t.Errorf("Render(invalid) = %q, want UTC result %q", got, wantUTC)
	}
}

// ---------------------------------------------------------------------------
// RenderWithFormat
// ---------------------------------------------------------------------------

func TestRenderWithFormat_UtcSecTsU32(t *testing.T) {
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
			got := testU32Ts.RenderWithFormat(tt.timezone, *gdf)
			if got != tt.expected {
				t.Errorf("RenderWithFormat(%q, %q) = %q, want %q", tt.timezone, tt.format, got, tt.expected)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// DST handling
// ---------------------------------------------------------------------------

func TestRenderWithFormat_UtcSecTsU32_DST(t *testing.T) {
	tests := []struct {
		name     string
		ts       UtcSecTsU32
		expected string
	}{
		{
			name:     "Winter (EST = UTC-5)",
			ts:       FromTimeU32(time.Date(2024, 1, 15, 12, 34, 56, 0, time.UTC)),
			expected: "2024-01-15 07:34:56 EST -0500",
		},
		{
			name:     "Summer (EDT = UTC-4)",
			ts:       FromTimeU32(time.Date(2024, 7, 15, 12, 34, 56, 0, time.UTC)),
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
// uint32 boundaries
// ---------------------------------------------------------------------------

// TestUtcSecTsU32_Beyond2038 verifies the uint32 type can represent dates
// beyond the Year 2038 problem that limits int32.
func TestUtcSecTsU32_Beyond2038(t *testing.T) {
	// 2100-01-01 00:00:00 UTC → Unix = 4102444800 (fits in uint32)
	ts := UtcSecTsU32(4_102_444_800)
	got := ts.String()
	want := "2100-01-01T00:00:00Z"
	if got != want {
		t.Errorf("Beyond2038: got %q, want %q", got, want)
	}
}

// TestUtcSecTsU32_MaxValue tests the uint32 maximum (2106-02-07 06:28:15 UTC).
func TestUtcSecTsU32_MaxValue(t *testing.T) {
	ts := UtcSecTsU32(0xFFFF_FFFF) // 4294967295
	got := ts.String()
	want := "2106-02-07T06:28:15Z"
	if got != want {
		t.Errorf("max uint32: got %q, want %q", got, want)
	}
}
