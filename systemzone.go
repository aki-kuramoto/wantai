package wantai

import "sync/atomic"

// SystemZoneName reports the IANA name of the zone this machine is set to
// ("Asia/Tokyo"), for recording next to an instant, for showing a person which
// zone a rendering used, or for passing to Render when a concrete name is
// wanted rather than "Local".
//
// The standard library has no equivalent: time.Local is a *time.Location that
// knows its offset but not its name, and time.Local.String() answers "Local"
// unless TZ happens to be set. So this reads what the platform recorded, and
// the answer agrees with time.Local on each platform:
//
//   - Unix: TZ when it is set, since Go honours it too; otherwise the name
//     embedded in /etc/localtime — normally a symlink into the zoneinfo tree,
//     and on systems where it is a plain copy instead, found by matching its
//     contents against that tree.
//   - Windows: the registry's TimeZoneKeyName, translated through CLDR.
//     Windows names its zones "Tokyo Standard Time", and Go ignores TZ there,
//     so neither TZ nor time.Local can answer this.
//
// The cost is wildly uneven — a few hundred nanoseconds when TZ is set, a few
// microseconds for a symlink, and milliseconds on a system whose /etc/localtime
// is a copy, because the whole zoneinfo tree then has to be read and compared.
// Which of those a machine is cannot be told from the code, so a successful
// answer is cached. Use [ClearSystemZoneNameCache] when the machine's zone may
// have changed. Concurrent first calls may each do the work; they agree, and
// paying twice once is cheaper than serialising every later call.
//
// An error means the zone could not be named, not that there is no zone: a
// container with a copied /etc/localtime and no matching file in its zoneinfo
// tree is a real configuration, and so is a Windows zone CLDR does not know.
// A caller that only needs to render can fall back to "Local", which needs no
// name.
func SystemZoneName() (string, error) {
	if cached, _ := systemZone.Load().(string); cached != "" {
		return cached, nil
	}
	name, err := systemZoneName()
	if err != nil {
		return "", err
	}
	systemZone.Store(name)
	return name, nil
}

// ClearSystemZoneNameCache drops the cached result of SystemZoneName, so the
// next call asks the platform again.
//
// It is deliberately not part of [ClearLocationCache]. The two caches answer
// different questions and go stale for different reasons: a resolved zone goes
// stale when the tz database is updated, this goes stale when the machine is
// set to a different zone. Someone who has just learned the machine moved
// should not have to throw away every parsed zone as well, and someone updating
// tzdata should not have to re-read /etc/localtime.
func ClearSystemZoneNameCache() {
	systemZone.Store("")
}

// systemZone caches the name. Always holds a string; the empty one means "not
// looked up yet", which is also what the clear stores.
var systemZone atomic.Value
