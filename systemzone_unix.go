//go:build !windows

package wantai

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// zoneInfoRoots are the places a zoneinfo tree is found, in the order the
// standard library looks. The macOS path is here and not there: Go reaches the
// same files through /usr/share/zoneinfo, which is a symlink, but the symlink
// /etc/localtime points at the real location, and a prefix has to match the
// text actually stored in it.
var zoneInfoRoots = []string{
	"/usr/share/zoneinfo/",
	"/usr/share/lib/zoneinfo/",
	"/usr/lib/locale/TZ/",
	"/etc/zoneinfo/",
	"/var/db/timezone/zoneinfo/", // macOS
}

const localtimePath = "/etc/localtime"

func systemZoneName() (string, error) {
	// TZ wins, because it wins for the standard library too: a process started
	// with TZ=Asia/Tokyo has time.Local in Tokyo whatever /etc/localtime says,
	// and this has to agree with time.Local or the two are answers to different
	// questions. The forms Go treats as UTC are reported as UTC.
	if tz, ok := os.LookupEnv("TZ"); ok {
		switch tz {
		case "", "UTC":
			return "UTC", nil
		}
		// A leading colon is allowed and ignored (POSIX); an absolute path is a
		// file rather than a name, so run it through the same prefix trimming.
		tz = strings.TrimPrefix(tz, ":")
		if strings.HasPrefix(tz, "/") {
			if name, ok := trimZoneRoot(tz); ok {
				return name, nil
			}
			return "", fmt.Errorf("wantai: TZ is %q, which is not inside a known zoneinfo directory", tz)
		}
		return tz, nil
	}

	// The common case: /etc/localtime is a symlink and the name is its target.
	if target, err := os.Readlink(localtimePath); err == nil {
		if !filepath.IsAbs(target) {
			target = filepath.Join(filepath.Dir(localtimePath), target)
		}
		if name, ok := trimZoneRoot(filepath.Clean(target)); ok {
			return name, nil
		}
		return "", fmt.Errorf("wantai: %s points at %q, which is not inside a known zoneinfo directory",
			localtimePath, target)
	}

	// Some systems — container images especially — copy the zone file instead
	// of linking it, and the name is simply gone. It can only be recovered by
	// finding the file in the tree with the same contents. This reads the whole
	// tree in the worst case, which is why SystemZoneName caches.
	want, err := os.ReadFile(localtimePath)
	if err != nil {
		return "", fmt.Errorf("wantai: cannot determine the system timezone: %w", err)
	}
	if name, ok := searchZoneInfo(want); ok {
		return name, nil
	}
	return "", errors.New("wantai: cannot determine the system timezone: " + localtimePath +
		" is a copy rather than a symlink, and no file in the zoneinfo tree matches it")
}

// trimZoneRoot turns an absolute path inside a zoneinfo tree into the zone name.
func trimZoneRoot(path string) (string, bool) {
	for _, root := range zoneInfoRoots {
		if strings.HasPrefix(path, root) {
			if name := strings.TrimPrefix(path, root); name != "" {
				return name, true
			}
		}
	}
	return "", false
}

// searchZoneInfo finds the zone whose file is byte-for-byte want.
//
// Ties are normal: zones that have always agreed share one file, so Asia/Tokyo
// and Japan are the same bytes, as are America/New_York and US/Eastern. Every
// matching name describes the same instants, so no choice here is wrong — but
// the canonical spelling is the one worth handing back, and which of the two it
// is cannot be read from the files. It is inferred instead, by preferring the
// Area/Location layout the tz database uses for current zones over the
// single-word and country-prefixed names it keeps for compatibility.
//
// This is a heuristic and it is not perfect: Australia/ACT and Australia/Sydney
// are both under a real area, and the tie-break picks the shorter one. Both
// name the same zone, so a caller renders the same instant either way; only the
// spelling differs. Exactness would mean carrying tzdata's backward file, which
// is not worth it for a branch this rare.
func searchZoneInfo(want []byte) (string, bool) {
	best := ""
	for _, root := range zoneInfoRoots {
		info, err := os.Stat(root)
		if err != nil || !info.IsDir() {
			continue
		}
		_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			name := strings.TrimPrefix(path, root)
			// posixrules and the like are not zones anyone means.
			if name == "" || strings.HasPrefix(name, "posix/") || strings.HasPrefix(name, "right/") {
				return nil
			}
			if best != "" && !betterZoneName(name, best) {
				return nil
			}
			got, err := os.ReadFile(path)
			if err != nil || !bytes.Equal(got, want) {
				return nil
			}
			best = name
			return nil
		})
		if best != "" {
			return best, true
		}
	}
	return "", false
}

// canonicalAreas are the top-level directories the tz database uses for zones
// that are current. Anything else at the top level — Japan, Egypt, US/Eastern,
// Canada/Atlantic — is kept for compatibility and names a zone that also has an
// Area/Location spelling.
var canonicalAreas = []string{
	"Africa/", "America/", "Antarctica/", "Arctic/", "Asia/", "Atlantic/",
	"Australia/", "Europe/", "Indian/", "Pacific/", "Etc/",
}

// betterZoneName reports whether a should win over b as the name for a zone
// file that both describe. An Area/Location name beats one that is not; then
// the shorter; then alphabetical order, so the answer does not depend on the
// order the tree happens to be walked in.
func betterZoneName(a, b string) bool {
	if ca, cb := inCanonicalArea(a), inCanonicalArea(b); ca != cb {
		return ca
	}
	if len(a) != len(b) {
		return len(a) < len(b)
	}
	return a < b
}

func inCanonicalArea(name string) bool {
	for _, area := range canonicalAreas {
		if strings.HasPrefix(name, area) {
			return true
		}
	}
	return false
}
