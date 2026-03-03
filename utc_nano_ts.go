package wantai

import (
	"sync"
	"time"
)

// UtcNanoTs represents a UTC nanosecond Unix timestamp.
// It can store the value of time.Now().UnixNano() directly.
type UtcNanoTs int64

// Zero is a UtcNanoTs representing the Unix epoch (1970-01-01T00:00:00Z).
const Zero = UtcNanoTs(0)

// locEntry caches a resolved timezone together with its DST status and fixed offset.
type locEntry struct {
	loc       *time.Location
	hasDST    bool   // true if the timezone observes Daylight Saving Time
	tzName    string // standard (non-DST) name; valid only when hasDST == false
	offsetSec int    // standard offset in seconds; valid only when hasDST == false
}

var locationCache sync.Map // map[string]locEntry

// FromTime converts a time.Time to UtcNanoTs.
// The result is always stored as a UTC nanosecond Unix timestamp,
// regardless of the timezone of the input time.Time.
func FromTime(t time.Time) UtcNanoTs {
	return UtcNanoTs(t.UnixNano())
}

// ToTime converts UtcNanoTs to a time.Time in UTC.
func (ts UtcNanoTs) ToTime() time.Time {
	return time.Unix(0, int64(ts)).UTC()
}

// String implements fmt.Stringer.
// It returns the timestamp formatted as RFC3339 in UTC.
func (ts UtcNanoTs) String() string {
	return ts.Render("UTC")
}

// Render returns the timestamp formatted as RFC3339 in the given timezone.
// If the timezone name is invalid, it falls back to UTC.
func (ts UtcNanoTs) Render(timezone string) string {
	entry, err := getLocEntryWithCache(timezone)
	loc := time.UTC
	if err == nil {
		loc = entry.loc
	}
	sec := int64(ts) / 1_000_000_000
	nano := int64(ts) % 1_000_000_000
	if nano < 0 {
		sec--
		nano += 1_000_000_000
	}
	return time.Unix(sec, nano).In(loc).Format(time.RFC3339)
}

// RenderWithFormat returns the timestamp rendered using the given timezone and GeneralDateFormat.
// If the timezone name is invalid, it falls back to UTC.
//
// For timezones without DST (e.g. UTC, Asia/Tokyo), all fields are computed via pure
// integer arithmetic — no time.Time is created at render time.
// For timezones with DST (e.g. America/New_York), time.Time field accessors are used
// to obtain the correct UTC offset for the given instant.
func (ts UtcNanoTs) RenderWithFormat(timezone string, format GeneralDateFormat) string {
	entry, err := getLocEntryWithCache(timezone)
	if err != nil {
		// Fall back to UTC on invalid timezone.
		entry, _ = getLocEntryWithCache("UTC")
	}

	// Split timestamp into whole seconds and sub-second nanoseconds.
	unixSec := int64(ts) / 1_000_000_000
	absNano := int64(ts) % 1_000_000_000
	if absNano < 0 {
		unixSec--
		absNano += 1_000_000_000
	}

	if !entry.hasDST {
		// Pure arithmetic path for fixed-offset timezones.
		adjustedSec := unixSec + int64(entry.offsetSec)
		return format.render(adjustedSec, absNano, entry.offsetSec, entry.tzName)
	}

	// DST path: use time.Time to get the correct offset for this instant.
	// We do NOT call t.Format(); we only read the offset via t.Zone().
	t := time.Unix(unixSec, absNano).In(entry.loc)
	tzName, offsetSec := t.Zone()
	adjustedSec := unixSec + int64(offsetSec)
	return format.render(adjustedSec, absNano, offsetSec, tzName)
}

// getLocEntryWithCache loads and caches a locEntry for the given timezone name.
func getLocEntryWithCache(locationName string) (locEntry, error) {
	if v, ok := locationCache.Load(locationName); ok {
		return v.(locEntry), nil
	}
	loc, err := time.LoadLocation(locationName)
	if err != nil {
		return locEntry{}, err
	}

	// Detect DST by sampling offsets at four points throughout a reference year
	// (roughly every 3 months). Comparing only two points (e.g. Jan and Jul) fails
	// for Southern Hemisphere zones where both those instants may fall in standard
	// time. Four samples covering all four quarters reliably catch any DST.
	hasDST := detectDST(loc)

	var entry locEntry
	if hasDST {
		entry = locEntry{loc: loc, hasDST: true}
	} else {
		tzName, offsetSec := time.Unix(0, 0).In(loc).Zone()
		entry = locEntry{loc: loc, hasDST: false, tzName: tzName, offsetSec: offsetSec}
	}

	locationCache.Store(locationName, entry)
	return entry, nil
}

// ClearLocationCache clears all cached timezone locations.
// Call this after updating the timezone database, or in tests that require a fresh cache.
func ClearLocationCache() {
	locationCache = sync.Map{}
}

// detectDST reports whether loc observes Daylight Saving Time.
//
// Strategy: sample 12 instants at monthly intervals around the current time.
// If any two consecutive samples have different UTC offsets, the timezone has DST.
// Sampling relative to the current time avoids issues with historical timezone
// rules that differ from modern behavior (e.g. Southern Hemisphere zones in 1970).
func detectDST(loc *time.Location) bool {
	now := time.Now().Unix()
	const month = 30 * 24 * 3600
	_, ref := time.Unix(now, 0).In(loc).Zone()
	for i := int64(1); i <= 12; i++ {
		_, off := time.Unix(now+i*month, 0).In(loc).Zone()
		if off != ref {
			return true
		}
	}
	return false
}
