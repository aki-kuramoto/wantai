package wantai

import (
	"fmt"
	"math"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// boundaryFormat is the rich format used for all boundary-value tests.
// It exercises date, time, sub-second, timezone name, and offset tokens.
// ---------------------------------------------------------------------------

const boundaryFormat = "YYYY-MM-DD HH:mm:ss.SSSfffnnn Z ZZ"

// ---------------------------------------------------------------------------
// renderExpected computes the expected render output using time.Time field
// accessors (NOT time.Format) as the authoritative reference.
// It is shared by boundary tests for every timestamp type.
// ---------------------------------------------------------------------------

func renderExpected(sec, nano int64, loc *time.Location) string {
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

// ---------------------------------------------------------------------------
// Boundary timestamp sets (UtcNanoTs, shared across files)
// ---------------------------------------------------------------------------

// boundaryTimestamps1677 covers the area around the int64 nanosecond minimum
// (≈ 1677-09-21 00:12:43 UTC).
var boundaryTimestamps1677 = []UtcNanoTs{
	UtcNanoTs(math.MinInt64),                         // 1677-09-21 minimum
	UtcNanoTs(math.MinInt64 + 1),                     // +1 ns
	UtcNanoTs(math.MinInt64 + 1_000_000_000),         // +1 s
	UtcNanoTs(math.MinInt64 + 86400*1_000_000_000),   // +1 day (~1677-09-22)
	UtcNanoTs(math.MinInt64 + 7*86400*1_000_000_000), // +7 days
}

// boundaryTimestamps2262 covers the area around the int64 nanosecond maximum
// (≈ 2262-04-11 23:47:16 UTC).
var boundaryTimestamps2262 = []UtcNanoTs{
	UtcNanoTs(math.MaxInt64),                         // 2262-04-11 maximum
	UtcNanoTs(math.MaxInt64 - 1),                     // -1 ns
	UtcNanoTs(math.MaxInt64 - 1_000_000_000),         // -1 s
	UtcNanoTs(math.MaxInt64 - 86400*1_000_000_000),   // -1 day (~2262-04-10)
	UtcNanoTs(math.MaxInt64 - 7*86400*1_000_000_000), // -7 days
}

// runNanoBoundaryTests is the boundary test runner for UtcNanoTs.
func runNanoBoundaryTests(t *testing.T, timezone string, tss []UtcNanoTs) {
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
		// Decompose nanosecond timestamp for renderExpected.
		rawNano := int64(ts)
		sec := rawNano / 1_000_000_000
		nano := rawNano % 1_000_000_000
		if nano < 0 {
			sec--
			nano += 1_000_000_000
		}
		want := renderExpected(sec, nano, entry.loc)
		if got != want {
			t.Errorf("tz=%q ts=%d:\n got  %q\n want %q", timezone, rawNano, got, want)
		}
	}
}
