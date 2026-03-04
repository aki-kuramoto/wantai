package wantai

import "time"

// UtcMicroTs represents a UTC microsecond Unix timestamp.
type UtcMicroTs int64

// FromTimeMicros converts a time.Time to UtcMicroTs.
// The result is always stored as a UTC microsecond Unix timestamp,
// regardless of the timezone of the input time.Time.
func FromTimeMicros(t time.Time) UtcMicroTs {
	return UtcMicroTs(t.UnixMicro())
}

// ToTime converts UtcMicroTs to a time.Time in UTC.
func (ts UtcMicroTs) ToTime() time.Time {
	us := int64(ts)
	sec := us / 1_000_000
	nano := (us % 1_000_000) * 1_000
	if nano < 0 {
		sec--
		nano += 1_000_000_000
	}
	return time.Unix(sec, nano).UTC()
}

// String implements fmt.Stringer.
// It returns the timestamp formatted as RFC3339 in UTC.
func (ts UtcMicroTs) String() string {
	return ts.Render("UTC")
}

// Render returns the timestamp formatted as RFC3339 in the given timezone.
// If the timezone name is invalid, it falls back to UTC.
func (ts UtcMicroTs) Render(timezone string) string {
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
func (ts UtcMicroTs) RenderWithFormat(timezone string, format GeneralDateFormat) string {
	entry, err := getLocEntryWithCache(timezone)
	if err != nil {
		entry, _ = getLocEntryWithCache("UTC")
	}

	// Convert microseconds to whole seconds and sub-second nanoseconds.
	us := int64(ts)
	unixSec := us / 1_000_000
	absNano := (us % 1_000_000) * 1_000
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
