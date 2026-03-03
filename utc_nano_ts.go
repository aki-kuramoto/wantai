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

var locationCache sync.Map

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
	return ts.RenderWithGoLayout(timezone, time.RFC3339)
}

// RenderWithFormat returns the timestamp formatted using the given timezone and GeneralDateFormat.
// If the timezone name is invalid, it falls back to UTC.
func (ts UtcNanoTs) RenderWithFormat(timezone string, format GeneralDateFormat) string {
	goLayout := format.GoLayout()
	return ts.RenderWithGoLayout(timezone, goLayout)
}

// RenderWithGoLayout returns the timestamp formatted using the given timezone and Go layout string.
// If the timezone name is invalid, it falls back to UTC.
func (ts UtcNanoTs) RenderWithGoLayout(timezone string, layout string) string {
	loc, err := getLocationWithCache(timezone)
	if err != nil {
		loc = time.UTC
	}
	t := time.Unix(0, int64(ts)).In(loc)
	return t.Format(layout)
}

func getLocationWithCache(locationName string) (*time.Location, error) {
	if loc, ok := locationCache.Load(locationName); ok {
		return loc.(*time.Location), nil
	}
	loc, err := time.LoadLocation(locationName)
	if err == nil {
		locationCache.Store(locationName, loc)
	}
	return loc, err
}

// ClearLocationCache clears all cached timezone locations.
// Call this after updating the timezone database, or in tests that require a fresh cache.
func ClearLocationCache() {
	locationCache = sync.Map{}
}
