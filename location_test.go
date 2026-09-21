package wantai

import (
	"sync"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// ClearLocationCache
// ---------------------------------------------------------------------------

func TestClearLocationCache(t *testing.T) {
	// Populate the cache without depending on any specific timestamp type.
	_, err := getLocEntryWithCache("Asia/Tokyo")
	if err != nil {
		t.Fatalf("getLocEntryWithCache: %v", err)
	}
	if _, ok := locationCache.Load("Asia/Tokyo"); !ok {
		t.Fatal("Expected Asia/Tokyo to be cached")
	}

	ClearLocationCache()

	if _, ok := locationCache.Load("Asia/Tokyo"); ok {
		t.Error("Expected cache to be empty after ClearLocationCache()")
	}
}

// ---------------------------------------------------------------------------
// detectDST
// ---------------------------------------------------------------------------

func TestDetectDST_HasDST(t *testing.T) {
	zones := []string{
		"America/New_York",
		"Europe/London",
		"Australia/Sydney",
		"Pacific/Auckland",
	}
	for _, tz := range zones {
		loc, err := loadLoc(t, tz)
		if err != nil {
			continue
		}
		if !detectDST(loc) {
			t.Errorf("detectDST(%q) = false, want true", tz)
		}
	}
}

func TestDetectDST_NoDST(t *testing.T) {
	zones := []string{
		"UTC",
		"Asia/Tokyo",
		"Asia/Kolkata",
		"Africa/Nairobi",
	}
	for _, tz := range zones {
		loc, err := loadLoc(t, tz)
		if err != nil {
			continue
		}
		if detectDST(loc) {
			t.Errorf("detectDST(%q) = true, want false", tz)
		}
	}
}

// ---------------------------------------------------------------------------
// getLocEntryWithCache
// ---------------------------------------------------------------------------

func TestGetLocEntryWithCache_Invalid(t *testing.T) {
	_, err := getLocEntryWithCache("Invalid/Timezone")
	if err == nil {
		t.Error("expected error for invalid timezone, got nil")
	}
}

func TestGetLocEntryWithCache_CachesResult(t *testing.T) {
	ClearLocationCache()
	_, _ = getLocEntryWithCache("Asia/Tokyo")
	if _, ok := locationCache.Load("Asia/Tokyo"); !ok {
		t.Error("expected Asia/Tokyo to be stored in cache after first call")
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func loadLoc(t *testing.T, tz string) (*time.Location, error) {
	t.Helper()
	ClearLocationCache()
	entry, err := getLocEntryWithCache(tz)
	if err != nil {
		t.Logf("skipping %q: %v", tz, err)
		return nil, err
	}
	return entry.loc, nil
}

// ---------------------------------------------------------------------------
// TryToLoadZoneForErrorCheck
// ---------------------------------------------------------------------------

// The reason TryToLoadZoneForErrorCheck exists: the renderers answer an unresolvable name with
// UTC, which is a well-formed timestamp in the wrong zone and is indistinguish-
// able from having asked for UTC. These are the names a person actually types.
func TestLoadZoneRejectsWhatTheRenderersAccept(t *testing.T) {
	ts := FromTimeMillis(time.Date(2024, 1, 15, 12, 34, 56, 0, time.UTC))
	utc := ts.Render("UTC")

	// Deliberately not "GMT+9": Etc/GMT+9 exists in the tz database and whether
	// a bare "GMT+9" resolves is a packaging detail that varies by platform.
	// These four do not resolve anywhere.
	for _, name := range []string{"JST", "PST", "+09:00", "Nonsense/Zone"} {
		t.Run(name, func(t *testing.T) {
			if err := TryToLoadZoneForErrorCheck(name); err == nil {
				t.Errorf("TryToLoadZoneForErrorCheck(%q) accepted a name the tz database does not know", name)
			}
			// And this is what it would have rendered as: UTC, with nothing to
			// distinguish it from the real thing.
			if got := ts.Render(name); got != utc {
				t.Errorf("Render(%q) = %q, want the UTC fallback %q", name, got, utc)
			}
		})
	}
}

func TestLoadZoneAcceptsTheNamesThatWork(t *testing.T) {
	for _, name := range []string{"", "UTC", "Local", "Asia/Tokyo", "America/New_York"} {
		label := name
		if label == "" {
			label = "(empty)"
		}
		t.Run(label, func(t *testing.T) {
			if err := TryToLoadZoneForErrorCheck(name); err != nil {
				t.Errorf("TryToLoadZoneForErrorCheck(%q): %v", name, err)
			}
		})
	}
}

// "Local" is the documented answer to "render in the zone this process runs in",
// so it has to actually resolve to that zone rather than quietly meaning UTC.
func TestLocalNamesTheProcessZone(t *testing.T) {
	old := time.Local
	time.Local = time.FixedZone("TEST", 5*60*60)
	ClearLocationCache() // the name is cached, so the pin needs a fresh lookup
	t.Cleanup(func() { time.Local = old; ClearLocationCache() })

	ts := FromTimeMillis(time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC))
	if got, want := ts.Render("Local"), "2024-01-15T17:00:00+05:00"; got != want {
		t.Errorf("Render(%q) = %q, want %q", "Local", got, want)
	}
	if got := ts.Render("UTC"); got == ts.Render("Local") {
		t.Errorf("Local rendered the same as UTC (%q) — it did not resolve to the process zone", got)
	}
}

// A successful TryToLoadZoneForErrorCheck leaves the zone cached, so the first render is not the
// one that pays for the lookup.
func TestLoadZoneCaches(t *testing.T) {
	ClearLocationCache()
	t.Cleanup(ClearLocationCache)

	if _, ok := locationCache.Load("Europe/Berlin"); ok {
		t.Fatal("the cache should be empty at the start of this test")
	}
	if err := TryToLoadZoneForErrorCheck("Europe/Berlin"); err != nil {
		t.Fatal(err)
	}
	if _, ok := locationCache.Load("Europe/Berlin"); !ok {
		t.Error("TryToLoadZoneForErrorCheck resolved the name but did not cache it")
	}
}

// ClearLocationCache used to replace the package variable with a fresh
// sync.Map. sync.Map guards its own contents, but assigning to the variable is
// a write to the variable, and a renderer on another goroutine is reading it at
// the same time. go vet stays quiet (a new composite literal is not a copied
// lock), so only -race catches it. Run this with -race.
func TestClearLocationCacheIsSafeWhileRendering(t *testing.T) {
	ts := FromTimeMillis(time.Now())
	stop := make(chan struct{})

	var renderers sync.WaitGroup
	for i := 0; i < 4; i++ {
		renderers.Add(1)
		go func() {
			defer renderers.Done()
			for {
				select {
				case <-stop:
					return
				default:
					_ = ts.Render("Asia/Tokyo")
				}
			}
		}()
	}

	var clearer sync.WaitGroup
	clearer.Add(1)
	go func() {
		defer clearer.Done()
		for i := 0; i < 500; i++ {
			ClearLocationCache()
		}
	}()

	clearer.Wait() // the clearing is bounded; the renderers only stop on demand
	close(stop)
	renderers.Wait()
}

// ---------------------------------------------------------------------------
// ToTimeIn
// ---------------------------------------------------------------------------

func TestToTimeInResolvesTheZone(t *testing.T) {
	ts := FromTimeMillis(time.Date(2024, 1, 15, 12, 34, 56, 0, time.UTC))

	got, err := ts.ToTimeIn("Asia/Tokyo")
	if err != nil {
		t.Fatalf("ToTimeIn: %v", err)
	}
	// Same instant, read in another zone.
	if !got.Equal(ts.ToTime()) {
		t.Errorf("the instant moved: %s vs %s", got, ts.ToTime())
	}
	if h := got.Hour(); h != 21 {
		t.Errorf("hour = %d, want 21 (12:34 UTC is 21:34 in Tokyo)", h)
	}
	if _, off := got.Zone(); off != 9*60*60 {
		t.Errorf("offset = %d, want %d", off, 9*60*60)
	}
}

// The point of returning an error: an unresolvable name must not come back as a
// plausible instant in the wrong zone, the way Render's UTC fallback does.
func TestToTimeInReportsABadZoneInsteadOfGuessing(t *testing.T) {
	ts := FromTimeMillis(time.Date(2024, 1, 15, 12, 34, 56, 0, time.UTC))

	got, err := ts.ToTimeIn("JST")
	if err == nil {
		t.Fatal("ToTimeIn accepted \"JST\", which is an abbreviation and not a zone name")
	}
	if !got.IsZero() {
		t.Errorf("got %s, want the zero time — a caller who ignores the error should "+
			"get something obviously broken, not a wrong-but-plausible instant", got)
	}
}

// DST is the case a fixed offset would get wrong, so check both sides of it.
func TestToTimeInFollowsDST(t *testing.T) {
	summer := FromTimeMillis(time.Date(2024, 7, 1, 16, 0, 0, 0, time.UTC))
	winter := FromTimeMillis(time.Date(2024, 1, 1, 16, 0, 0, 0, time.UTC))

	s, err := summer.ToTimeIn("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	w, err := winter.ToTimeIn("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	if _, off := s.Zone(); off != -4*60*60 {
		t.Errorf("July offset = %d, want %d (EDT)", off, -4*60*60)
	}
	if _, off := w.Zone(); off != -5*60*60 {
		t.Errorf("January offset = %d, want %d (EST)", off, -5*60*60)
	}
}

// "Local" works here for the same reason it works in Render.
func TestToTimeInAcceptsLocal(t *testing.T) {
	old := time.Local
	time.Local = time.FixedZone("TEST", 5*60*60)
	ClearLocationCache()
	t.Cleanup(func() { time.Local = old; ClearLocationCache() })

	ts := FromTimeMillis(time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC))
	got, err := ts.ToTimeIn("Local")
	if err != nil {
		t.Fatal(err)
	}
	if h := got.Hour(); h != 17 {
		t.Errorf("hour = %d, want 17", h)
	}
}
