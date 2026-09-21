package wantai

import "time"

// UtcMilliTs represents a UTC millisecond Unix timestamp.
// It can store the value of time.Now().UnixMilli() directly.
type UtcMilliTs int64

// FromTimeMillis converts a time.Time to UtcMilliTs.
// The result is always stored as a UTC millisecond Unix timestamp,
// regardless of the timezone of the input time.Time.
func FromTimeMillis(t time.Time) UtcMilliTs {
	return UtcMilliTs(t.UnixMilli())
}

// ToTime converts UtcMilliTs to a time.Time in UTC.
func (ts UtcMilliTs) ToTime() time.Time {
	ms := int64(ts)
	sec := ms / 1_000
	nano := (ms % 1_000) * 1_000_000
	if nano < 0 {
		sec--
		nano += 1_000_000_000
	}
	return time.Unix(sec, nano).UTC()
}

// ToTimeIn converts UtcMilliTs to a time.Time in the given timezone.
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
func (ts UtcMilliTs) ToTimeIn(timezone string) (time.Time, error) {
	entry, err := getLocEntryWithCache(timezone)
	if err != nil {
		return time.Time{}, err
	}
	return ts.ToTime().In(entry.loc), nil
}

// String implements fmt.Stringer.
// It returns the timestamp formatted as RFC3339 in UTC.
func (ts UtcMilliTs) String() string {
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
func (ts UtcMilliTs) Render(timezone string) string {
	entry, err := getLocEntryWithCache(timezone)
	loc := time.UTC
	if err == nil {
		loc = entry.loc
	}
	return ts.ToTime().In(loc).Format(time.RFC3339)
}

// RenderWithFormat returns the timestamp rendered using the given timezone and GeneralDateFormat.
//
// timezone follows the same rules as [Render]'s: an IANA name, or "", "UTC" or
// "Local"; an unresolvable one renders as UTC without saying so. See the package
// documentation, and [TryToLoadZoneForErrorCheck] for checking a name before it
// gets here.
//
// For timezones without DST (e.g. UTC, Asia/Tokyo), all fields are computed via pure
// integer arithmetic — no time.Time is created at render time.
// For timezones with DST (e.g. America/New_York), time.Time field accessors are used
// to obtain the correct UTC offset for the given instant.
func (ts UtcMilliTs) RenderWithFormat(timezone string, format GeneralDateFormat) string {
	entry, err := getLocEntryWithCache(timezone)
	if err != nil {
		entry, _ = getLocEntryWithCache("UTC")
	}

	// Convert milliseconds to whole seconds and sub-second nanoseconds.
	ms := int64(ts)
	unixSec := ms / 1_000
	absNano := (ms % 1_000) * 1_000_000
	if absNano < 0 {
		unixSec--
		absNano += 1_000_000_000
	}

	if !entry.hasDST {
		adjustedSec := unixSec + int64(entry.offsetSec)
		return format.render(adjustedSec, absNano, entry.offsetSec, entry.tzName)
	}

	t := time.Unix(unixSec, absNano).In(entry.loc)
	tzName, offsetSec := t.Zone()
	adjustedSec := unixSec + int64(offsetSec)
	return format.render(adjustedSec, absNano, offsetSec, tzName)
}
