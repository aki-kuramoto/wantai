//go:build windows && !wantai_no_tzdata_on_windows

// The GOOS constraint is stated twice on purpose — once in the file name and
// once above. The file name alone would constrain the build, but then the line
// above would read as if it applied everywhere, and this file is precisely
// about one platform.

package wantai

// Windows has no system timezone database to read.
//
// Go's time package says so plainly: on Windows, platformZoneSources is empty
// ("none: Windows uses system calls instead"). Those system calls give the
// current zone's offset, not the tz database, so time.LoadLocation has no file
// source at all and an IANA name resolves only from an embedded copy, from
// ZONEINFO, or from $GOROOT/lib/time/zoneinfo.zip — that last one meaning the
// machine has Go installed, which a machine running a shipped binary does not.
//
// Without a copy, then, every name in this package's care falls back to UTC on
// Windows: a well-formed timestamp in the wrong zone, which is the failure this
// package exists to prevent. SystemZoneName makes it worse than a gap — it
// reads the zone from the registry and answers "Asia/Tokyo" without needing tz
// data at all, and Render would then refuse the very name it was handed.
//
// So the copy is embedded here, and the decision is deliberate. The standard
// library advises that a library should not make it ("Libraries normally
// shouldn't decide whether to include the timezone database in a program"),
// and that advice assumes the program can get the data elsewhere. On Windows it
// cannot.
//
// It costs about 400KB of binary size, measured. A program that renders only in
// "UTC" and "Local" does not need any of it — time.LoadLocation answers both
// without touching the database — and can build with
//
//	-tags wantai_no_tzdata_on_windows
//
// to leave it out. The tag does nothing on any other platform, where this file
// is not compiled.
//
// That name is long on purpose. wantai_no_windows_tzdata is shorter and reads
// as "leave out the Windows zone mapping" — a different table this package
// really does carry, in windowszones.go — while on_windows scopes the tag
// instead of appearing to name its contents. Build tags share one namespace
// across a whole build, which is what the wantai_ prefix is for.
//
// Go does not check the spelling of a build tag: an unknown one is ignored
// without a word, so a typo leaves the data in. That is the safe direction to
// fail — a binary larger than intended rather than one that reports the wrong
// time.
//
// -tags timetzdata is the standard library's own switch for the same data. A
// build that sets it gets the embedded copy whether or not this file is
// compiled.
import _ "time/tzdata"
