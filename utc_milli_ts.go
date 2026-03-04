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

// String implements fmt.Stringer.
// It returns the timestamp formatted as RFC3339 in UTC.
func (ts UtcMilliTs) String() string {
	return ts.Render("UTC")
}

// Render returns the timestamp formatted as RFC3339 in the given timezone.
// If the timezone name is invalid, it falls back to UTC.
func (ts UtcMilliTs) Render(timezone string) string {
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
