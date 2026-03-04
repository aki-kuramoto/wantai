package wantai

import "time"

// UtcSecTsU32 represents a UTC second Unix timestamp stored as an unsigned 32-bit integer.
// The representable range is 1970-01-01T00:00:00Z through 2106-02-07T06:28:15Z.
//
// Because uint32 cannot represent negative Unix timestamps, instants before the
// Unix epoch (1970-01-01) are not supported by this type.
type UtcSecTsU32 uint32

// FromTimeU32 converts a time.Time to UtcSecTsU32.
// Values outside the uint32 range (before epoch or after 2106) will overflow silently.
func FromTimeU32(t time.Time) UtcSecTsU32 {
	return UtcSecTsU32(uint32(t.Unix()))
}

// ToTime converts UtcSecTsU32 to a time.Time in UTC.
func (ts UtcSecTsU32) ToTime() time.Time {
	return time.Unix(int64(ts), 0).UTC()
}

// String implements fmt.Stringer.
// It returns the timestamp formatted as RFC3339 in UTC.
func (ts UtcSecTsU32) String() string {
	return ts.Render("UTC")
}

// Render returns the timestamp formatted as RFC3339 in the given timezone.
// If the timezone name is invalid, it falls back to UTC.
func (ts UtcSecTsU32) Render(timezone string) string {
	entry, err := getLocEntryWithCache(timezone)
	loc := time.UTC
	if err == nil {
		loc = entry.loc
	}
	return ts.ToTime().In(loc).Format(time.RFC3339)
}

// RenderWithFormat returns the timestamp rendered using the given timezone and GeneralDateFormat.
// If the timezone name is invalid, it falls back to UTC.
//
// For timezones without DST (e.g. UTC, Asia/Tokyo), all fields are computed via pure
// integer arithmetic — no time.Time is created at render time.
// For timezones with DST (e.g. America/New_York), time.Time field accessors are used
// to obtain the correct UTC offset for the given instant.
func (ts UtcSecTsU32) RenderWithFormat(timezone string, format GeneralDateFormat) string {
	entry, err := getLocEntryWithCache(timezone)
	if err != nil {
		entry, _ = getLocEntryWithCache("UTC")
	}

	unixSec := int64(ts)
	const absNano = int64(0) // second-level precision; no sub-second component

	if !entry.hasDST {
		adjustedSec := unixSec + int64(entry.offsetSec)
		return format.render(adjustedSec, absNano, entry.offsetSec, entry.tzName)
	}

	t := time.Unix(unixSec, 0).In(entry.loc)
	tzName, offsetSec := t.Zone()
	adjustedSec := unixSec + int64(offsetSec)
	return format.render(adjustedSec, absNano, offsetSec, tzName)
}
