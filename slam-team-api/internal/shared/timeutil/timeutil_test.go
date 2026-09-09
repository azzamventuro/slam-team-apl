package timeutil_test

import (
	"testing"
	"time"

	// Mirrors cmd/api and cmd/slamctl: the test must prove the zones resolve
	// from the embedded database, not from the host's tz files.
	_ "time/tzdata"

	"slam-team-api/internal/shared/timeutil"
)

func TestLocationResolvesIndonesianZones(t *testing.T) {
	for _, zone := range []string{"Asia/Jakarta", "Asia/Makassar", "Asia/Jayapura"} {
		if _, err := time.LoadLocation(zone); err != nil {
			t.Fatalf("LoadLocation(%q) = %v; tzdata not embedded?", zone, err)
		}
		if got := timeutil.Location(zone).String(); got != zone {
			t.Errorf("Location(%q) = %q", zone, got)
		}
	}
}

func TestLocationFallsBackToUTC(t *testing.T) {
	for _, zone := range []string{"", "Mars/Olympus"} {
		if got := timeutil.Location(zone); got != time.UTC {
			t.Errorf("Location(%q) = %v, want UTC", zone, got)
		}
	}
}

func TestInZoneKeepsTheInstant(t *testing.T) {
	utc := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	jkt := timeutil.InZone(utc, "Asia/Jakarta")

	if !jkt.Equal(utc) {
		t.Errorf("InZone changed the instant: %v != %v", jkt, utc)
	}
	if h := jkt.Hour(); h != 7 { // WIB is UTC+7
		t.Errorf("hour in Asia/Jakarta = %d, want 7", h)
	}
}
