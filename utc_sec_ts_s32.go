package wantai

import "time"

// UtcSecTsS32 represents a UTC second Unix timestamp stored as a signed 32-bit integer.
// The representable range is 1901-12-13T20:45:52Z through 2038-01-19T03:14:07Z.
//
// Note: this type is subject to the Year 2038 problem. For new systems,
// prefer UtcNanoTs (int64) or UtcSecTsU32 where the extended range suffices.
type UtcSecTsS32 int32

// FromTimeS32 converts a time.Time to UtcSecTsS32.
// Values outside the int32 range will overflow silently.
func FromTimeS32(t time.Time) UtcSecTsS32 {
	return UtcSecTsS32(int32(t.Unix()))
}

// ToTime converts UtcSecTsS32 to a time.Time in UTC.
func (ts UtcSecTsS32) ToTime() time.Time {
	return time.Unix(int64(ts), 0).UTC()
}

// String implements fmt.Stringer.
// It returns the timestamp formatted as RFC3339 in UTC.
func (ts UtcSecTsS32) String() string {
	return ts.Render("UTC")
}

// Render returns the timestamp formatted as RFC3339 in the given timezone.
// If the timezone name is invalid, it falls back to UTC.
func (ts UtcSecTsS32) Render(timezone string) string {
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
func (ts UtcSecTsS32) RenderWithFormat(timezone string, format GeneralDateFormat) string {
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
