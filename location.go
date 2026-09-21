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

// TryToLoadZoneForErrorCheck reports whether name is a timezone the renderers can resolve, and
// caches it if so.
//
// It exists because Render and RenderWithFormat cannot tell you. They return a
// string, and an unresolvable name renders as UTC — a well-formed timestamp that
// is simply in the wrong zone, indistinguishable from having asked for UTC on
// purpose. The names most likely to be typed by hand are exactly the ones that
// fail this way: "JST" and "+09:00" are not tz database names.
//
// So an application that takes a zone name from configuration, a user, or a
// database row should call TryToLoadZoneForErrorCheck once where it can still complain:
//
//	if err := wantai.TryToLoadZoneForErrorCheck(cfg.Timezone); err != nil {
//	    return fmt.Errorf("timezone %q: %w", cfg.Timezone, err)
//	}
//
// The accepted names are the ones [time.LoadLocation] accepts: an IANA name such
// as "Asia/Tokyo", or "", "UTC" and "Local" (see the package documentation).
//
// A successful call leaves the zone in the cache, so the first render does not
// pay for the lookup.
func TryToLoadZoneForErrorCheck(name string) error {
	_, err := getLocEntryWithCache(name)
	return err
}

// ClearLocationCache clears all cached timezone locations.
//
// Call it after updating the timezone database, and whenever the zone behind a
// name changes under a running process. The name "Local" is cached like any
// other, so a machine that changes zone — and a test that swaps [time.Local] —
// keeps rendering in the old one until the cache is dropped.
func ClearLocationCache() {
	// Delete the entries rather than replacing the map. Assigning a fresh
	// sync.Map to the package variable is a write to that variable, and a
	// renderer running on another goroutine is reading it at the same time —
	// a data race that -race reports and that go vet does not catch (a new
	// composite literal is not a copied lock, so copylocks stays quiet).
	// Range/Delete touches only the map's own synchronised state.
	// interface{} rather than any: go.mod declares go 1.17.
	locationCache.Range(func(k, _ interface{}) bool {
		locationCache.Delete(k)
		return true
	})
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
