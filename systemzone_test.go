package wantai

import (
	"runtime"
	"testing"
	"time"
)

// Whatever the platform answers, the answer has to be a name the renderers
// accept — that is the entire point of handing back an IANA name rather than an
// offset. A machine can legitimately fail to name its zone, and that is not a
// test failure; naming it wrongly is.
func TestSystemZoneNameIsUsableWhereNamesAreUsed(t *testing.T) {
	ClearSystemZoneNameCache()
	t.Cleanup(ClearSystemZoneNameCache)

	name, err := SystemZoneName()
	if err != nil {
		t.Skipf("this machine cannot name its zone, which is a supported outcome: %v", err)
	}
	t.Logf("system zone: %s", name)

	if err := TryToLoadZoneForErrorCheck(name); err != nil {
		t.Fatalf("SystemZoneName returned %q, which the renderers cannot resolve: %v", name, err)
	}

	// And it has to be the zone the process is actually in, or it is a name for
	// somewhere else. Compare offsets at a fixed instant rather than zone names,
	// which differ in spelling (JST vs Asia/Tokyo) by design.
	ts := FromTimeMillis(time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC))
	named, err := ts.ToTimeIn(name)
	if err != nil {
		t.Fatal(err)
	}
	local, err := ts.ToTimeIn("Local")
	if err != nil {
		t.Fatal(err)
	}
	_, namedOff := named.Zone()
	_, localOff := local.Zone()
	if namedOff != localOff {
		t.Errorf("SystemZoneName says %q (offset %d), but the process is at offset %d",
			name, namedOff, localOff)
	}
}

// TZ is what the standard library honours on Unix, so it has to be what this
// honours too — otherwise the name describes a different zone from the one
// time.Local is rendering in.
func TestSystemZoneNameFollowsTZ(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Go ignores TZ on Windows, and so does SystemZoneName")
	}
	for _, tc := range []struct{ tz, want string }{
		{"Asia/Tokyo", "Asia/Tokyo"},
		{"America/New_York", "America/New_York"},
		{"UTC", "UTC"},
		{"", "UTC"},
		{":Asia/Tokyo", "Asia/Tokyo"}, // POSIX allows the leading colon
	} {
		t.Run(tc.tz, func(t *testing.T) {
			t.Setenv("TZ", tc.tz)
			ClearSystemZoneNameCache()
			t.Cleanup(ClearSystemZoneNameCache)

			got, err := SystemZoneName()
			if err != nil {
				t.Fatalf("TZ=%q: %v", tc.tz, err)
			}
			if got != tc.want {
				t.Errorf("TZ=%q gave %q, want %q", tc.tz, got, tc.want)
			}
		})
	}
}

// The cache is the reason this is safe to call often, so it has to actually
// hold — and the clear has to actually clear.
func TestSystemZoneNameCacheAndItsClear(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("this drives the cache through TZ, which Windows does not use")
	}
	t.Setenv("TZ", "Asia/Tokyo")
	ClearSystemZoneNameCache()
	t.Cleanup(ClearSystemZoneNameCache)

	first, err := SystemZoneName()
	if err != nil {
		t.Fatal(err)
	}
	if first != "Asia/Tokyo" {
		t.Fatalf("got %q, want Asia/Tokyo", first)
	}

	// Change the source of truth without clearing: the cached answer must stand,
	// which is what "cached" means.
	t.Setenv("TZ", "America/New_York")
	if again, _ := SystemZoneName(); again != "Asia/Tokyo" {
		t.Errorf("got %q after changing TZ without clearing; the cache did not hold", again)
	}

	ClearSystemZoneNameCache()
	if after, err := SystemZoneName(); err != nil || after != "America/New_York" {
		t.Errorf("got %q (err %v) after clearing, want America/New_York", after, err)
	}
}

// Clearing the zone cache and clearing this one are separate on purpose, so
// neither may quietly do the other's job.
func TestTheTwoCachesAreIndependent(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("this drives the cache through TZ, which Windows does not use")
	}
	t.Setenv("TZ", "Asia/Tokyo")
	ClearSystemZoneNameCache()
	t.Cleanup(ClearSystemZoneNameCache)

	if _, err := SystemZoneName(); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TZ", "America/New_York")

	ClearLocationCache() // the other cache — must not disturb this one
	if got, _ := SystemZoneName(); got != "Asia/Tokyo" {
		t.Errorf("ClearLocationCache cleared the system zone name too (got %q)", got)
	}

	ClearSystemZoneNameCache()
	if got, _ := SystemZoneName(); got != "America/New_York" {
		t.Errorf("ClearSystemZoneNameCache did not clear (got %q)", got)
	}
}

// The CLDR table is only read on Windows, but it is data and can be checked
// anywhere — and this is the check that matters: every name it can produce has
// to be one the tz database actually has, or SystemZoneName hands back a name
// that fails at the next step.
func TestWindowsZonesMapToRealZones(t *testing.T) {
	if len(windowsZones) < 100 {
		t.Fatalf("the table has %d entries; the generator has probably broken", len(windowsZones))
	}
	for win, iana := range windowsZones {
		if err := TryToLoadZoneForErrorCheck(iana); err != nil {
			t.Errorf("%q maps to %q, which does not resolve: %v", win, iana, err)
		}
	}
}
