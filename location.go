package wantai

import (
	"sync"
	"time"
)

// locEntry caches a resolved timezone together with its DST status and fixed offset.
type locEntry struct {
	loc       *time.Location
	hasDST    bool   // true if the timezone observes Daylight Saving Time
	tzName    string // standard (non-DST) name; valid only when hasDST == false
	offsetSec int    // standard offset in seconds; valid only when hasDST == false
}

var locationCache sync.Map // map[string]locEntry

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
