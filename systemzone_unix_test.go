//go:build !windows

package wantai

import (
	"os"
	"path/filepath"
	"testing"
)

// The path SystemZoneName takes when /etc/localtime is a copy rather than a
// symlink — container images do this, and the name is simply gone, so it has to
// be recovered by matching contents. It is the expensive branch and the one a
// developer machine never exercises, so drive it directly.
func TestSearchZoneInfoRecoversTheName(t *testing.T) {
	// Pick a zone that exists wherever the tests run, and feed its own bytes
	// back in as if they had been copied to /etc/localtime.
	var want []byte
	var from string
	for _, root := range zoneInfoRoots {
		b, err := os.ReadFile(filepath.Join(root, "Asia/Tokyo"))
		if err == nil {
			want, from = b, root
			break
		}
	}
	if want == nil {
		t.Skip("no zoneinfo tree on this machine")
	}
	t.Logf("using the tree at %s", from)

	got, ok := searchZoneInfo(want)
	if !ok {
		t.Fatal("the contents of Asia/Tokyo were not found in the zoneinfo tree")
	}
	// Japan is a historical alias sharing the same file; either names the same
	// zone, but the shorter, canonical one is what the tie-break should pick.
	if got != "Asia/Tokyo" {
		t.Errorf("recovered %q, want Asia/Tokyo", got)
	}
}

func TestSearchZoneInfoReportsNoMatch(t *testing.T) {
	if _, ok := searchZoneInfo([]byte("not a zone file")); ok {
		t.Error("something matched bytes that are not a zone file")
	}
}

func TestTrimZoneRoot(t *testing.T) {
	for _, tc := range []struct {
		path, want string
		ok         bool
	}{
		{"/usr/share/zoneinfo/Asia/Tokyo", "Asia/Tokyo", true},
		{"/var/db/timezone/zoneinfo/Asia/Tokyo", "Asia/Tokyo", true}, // macOS
		{"/etc/zoneinfo/Europe/Berlin", "Europe/Berlin", true},
		{"/usr/share/zoneinfo/", "", false}, // a root with no zone after it
		{"/somewhere/else/Asia/Tokyo", "", false},
	} {
		got, ok := trimZoneRoot(tc.path)
		if got != tc.want || ok != tc.ok {
			t.Errorf("trimZoneRoot(%q) = %q, %v; want %q, %v", tc.path, got, ok, tc.want, tc.ok)
		}
	}
}

// A symlinked /etc/localtime is the common case; make sure the name comes out
// of it rather than out of TZ, which is what the earlier tests pinned.
func TestSystemZoneNameReadsTheSymlinkWhenTZIsUnset(t *testing.T) {
	if _, err := os.Readlink(localtimePath); err != nil {
		t.Skipf("%s is not a symlink here: %v", localtimePath, err)
	}
	os.Unsetenv("TZ")
	ClearSystemZoneNameCache()
	t.Cleanup(ClearSystemZoneNameCache)

	got, err := SystemZoneName()
	if err != nil {
		t.Fatal(err)
	}
	if err := TryToLoadZoneForErrorCheck(got); err != nil {
		t.Errorf("the symlink gave %q, which does not resolve: %v", got, err)
	}
	t.Logf("from the symlink: %s", got)
}
