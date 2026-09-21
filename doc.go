// Package wantai stores an instant as a UTC integer and renders a local reading
// from it on demand.
//
// The types (UtcNanoTs, UtcMicroTs, UtcMilliTs, UtcSecTsS32, UtcSecTsU32,
// UtcSecTsS32Ep2k, UtcSecTsU32Ep2k) are integers counting from an epoch, so a
// stored value carries no zone and cannot quietly become a local time the way a
// time.Time can. A zone enters the picture once, at the point of display, as the
// timezone argument to Render or RenderWithFormat.
//
// # Timezone names
//
// The timezone argument goes to [time.LoadLocation], so it is an IANA name from
// the tz database:
//
//	ts.Render("Asia/Tokyo")   // => "2024-01-15T21:34:56+09:00"
//
// plus the three names Go defines itself:
//
//	ts.Render("")             // UTC
//	ts.Render("UTC")          // UTC
//	ts.Render("Local")        // the zone this process is running in
//
// "Local" is the answer to "show it in the time of the machine this is running
// on". It needs no TZ environment variable and no lookup of the system's own
// name; [time.LoadLocation] documents it as the name that returns [time.Local].
//
// # A name is not an abbreviation and not an offset
//
// This is the sharp edge of the package. "JST" and "+09:00" are the two things a
// person is most likely to write, and neither is a name the tz database knows:
//
//	ts.Render("JST")          // => "2024-01-15T12:34:56Z"  — UTC, nine hours off
//	ts.Render("+09:00")       // => "2024-01-15T12:34:56Z"  — the same
//
// A name that cannot be resolved falls back to UTC, and the result is a
// well-formed timestamp that is simply wrong. Render returns only a string, so
// its result cannot tell you this happened: Render("UTC") and
// Render("Nonsense/Zone") are byte for byte identical.
//
// Spelling is part of the name. "asia/tokyo" resolves on macOS, whose filesystem
// does not distinguish case, and falls back to UTC on Linux — so a lowercase
// name can pass every test on a laptop and render UTC in production.
//
// Where a name comes from configuration, a user, or anywhere else it could be
// mistyped, check it once at startup with [TryToLoadZoneForErrorCheck] and fail
// loudly. Rendering
// is the wrong place to find out.
//
// # Naming the machine's own zone
//
// "Local" renders in the machine's zone without knowing its name. When the name
// itself is wanted — to record beside an instant, to show which zone a reading
// used, or because a concrete name is preferred to "Local" — [SystemZoneName]
// reports it. The standard library has no equivalent: time.Local knows its
// offset but not its name.
//
// # Caching
//
// A resolved zone is cached under the name it was asked for, so repeated
// rendering does not reload it. "Local" is a name like any other and is cached
// like any other: a process whose zone changes while it runs, and a test that
// swaps [time.Local], both keep seeing the old zone until [ClearLocationCache]
// is called.
package wantai
