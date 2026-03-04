package wantai

import (
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
