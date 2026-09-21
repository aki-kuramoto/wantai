package wantai

import "time"

// UtcNanoTs represents a UTC nanosecond Unix timestamp.
// It can store the value of time.Now().UnixNano() directly.
type UtcNanoTs int64

// Zero is a UtcNanoTs representing the Unix epoch (1970-01-01T00:00:00Z).
const Zero = UtcNanoTs(0)

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

// ToTimeIn converts UtcNanoTs to a time.Time in the given timezone.
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
func (ts UtcNanoTs) ToTimeIn(timezone string) (time.Time, error) {
	entry, err := getLocEntryWithCache(timezone)
	if err != nil {
		return time.Time{}, err
	}
	return ts.ToTime().In(entry.loc), nil
}

// String implements fmt.Stringer.
// It returns the timestamp formatted as RFC3339 in UTC.
func (ts UtcNanoTs) String() string {
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
