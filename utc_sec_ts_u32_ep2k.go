package wantai

import "time"

// UtcSecTsU32Ep2k represents a UTC second timestamp stored as an unsigned 32-bit integer
// with epoch at 2000-01-01T00:00:00Z (instead of the Unix epoch 1970-01-01T00:00:00Z).
//
// Representable range: 2000-01-01T00:00:00Z through 2136-02-07T06:28:15Z.
// Compared to UtcSecTsU32, the range starts 30 years later but also ends 30 years later,
// making it useful for systems that cannot represent pre-2000 timestamps anyway.
//
// Common use-case: embedded and IoT systems that define their own epoch at Y2K.
type UtcSecTsU32Ep2k uint32

// FromTimeU32Ep2k converts a time.Time to UtcSecTsU32Ep2k.
// Values outside the uint32 range relative to the Ep2k epoch will overflow silently.
// Times before 2000-01-01T00:00:00Z are not representable by this type.
func FromTimeU32Ep2k(t time.Time) UtcSecTsU32Ep2k {
	return UtcSecTsU32Ep2k(uint32(t.Unix() - epoch2kUnixSec))
}

// ToTime converts UtcSecTsU32Ep2k to a time.Time in UTC.
func (ts UtcSecTsU32Ep2k) ToTime() time.Time {
	return time.Unix(int64(ts)+epoch2kUnixSec, 0).UTC()
}

// String implements fmt.Stringer.
// It returns the timestamp formatted as RFC3339 in UTC.
func (ts UtcSecTsU32Ep2k) String() string {
	return ts.Render("UTC")
}

// Render returns the timestamp formatted as RFC3339 in the given timezone.
// If the timezone name is invalid, it falls back to UTC.
func (ts UtcSecTsU32Ep2k) Render(timezone string) string {
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
func (ts UtcSecTsU32Ep2k) RenderWithFormat(timezone string, format GeneralDateFormat) string {
	entry, err := getLocEntryWithCache(timezone)
	if err != nil {
		entry, _ = getLocEntryWithCache("UTC")
	}

	unixSec := int64(ts) + epoch2kUnixSec
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
