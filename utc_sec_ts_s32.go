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

// ToTimeIn converts UtcSecTsS32 to a time.Time in the given timezone.
//
// ToTime always answers in UTC, so a caller who wanted a local reading had to
// write ts.ToTime().In(loc) and resolve loc itself with time.LoadLocation —
// which is not cached by the standard library and costs roughly 12µs every
// call. ToTimeIn goes through the same cache the renderers use, so the same
// work is a few tens of nanoseconds after the first lookup.
//
// timezone follows the rules described in the package documentation: an IANA
// name, or "", "UTC" or "Local".
//
// Unlike Render, this reports an unresolvable name instead of falling back to
// UTC, and the returned time.Time is then the zero value rather than a
// plausible-looking instant in the wrong zone. Ignoring the error gives you
// something obviously broken, which is the intended trade.
func (ts UtcSecTsS32) ToTimeIn(timezone string) (time.Time, error) {
	entry, err := getLocEntryWithCache(timezone)
	if err != nil {
		return time.Time{}, err
	}
	return ts.ToTime().In(entry.loc), nil
}

// String implements fmt.Stringer.
// It returns the timestamp formatted as RFC3339 in UTC.
func (ts UtcSecTsS32) String() string {
	return ts.Render("UTC")
}

// Render returns the timestamp formatted as RFC3339 in the given timezone.
//
// timezone is an IANA name such as "Asia/Tokyo", or one of Go's own "", "UTC"
// and "Local" (the zone this process runs in). It is not an abbreviation and not
// an offset: "JST" and "+09:00" are not names, so they render as UTC — silently,
// and this result cannot tell you it happened. Validate with
// [TryToLoadZoneForErrorCheck] where
// the name did not come from a literal. See the package documentation.
func (ts UtcSecTsS32) Render(timezone string) string {
	entry, err := getLocEntryWithCache(timezone)
	loc := time.UTC
	if err == nil {
		loc = entry.loc
	}
	return ts.ToTime().In(loc).Format(time.RFC3339)
}

// RenderWithFormat returns the timestamp rendered using the given timezone and GeneralDateFormat.
//
// timezone follows the same rules as Render's: an IANA name, or "", "UTC" or
// "Local"; an unresolvable one renders as UTC without saying so. See the package
// documentation, and [TryToLoadZoneForErrorCheck] for checking a name before it
// gets here.
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
