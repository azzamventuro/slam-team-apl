package service

import (
	"testing"
	"time"

	"slam-team-api/internal/modules/core/jadwal/domain"
)

func d(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return t
}

func dp(s string) *time.Time { t := d(s); return &t }

func keys(ts []time.Time) []string {
	out := make([]string, len(ts))
	for i, t := range ts {
		out[i] = t.Format("2006-01-02")
	}
	return out
}

func assertDates(t *testing.T, got []time.Time, want ...string) {
	t.Helper()
	g := keys(got)
	if len(g) != len(want) {
		t.Fatalf("got %d dates %v, want %d %v", len(g), g, len(want), want)
	}
	for i := range g {
		if g[i] != want[i] {
			t.Fatalf("index %d: got %s, want %s (all: %v)", i, g[i], want[i], g)
		}
	}
}

func TestDatesTidakBerulang(t *testing.T) {
	rule := Rule{Pola: domain.PolaTidakBerulang, TanggalMulai: d("2026-09-15"), HariUlang: []int64{0, 1, 2}}
	// hari_ulang ignored; exactly one date.
	assertDates(t, Dates(rule, d("2026-09-01"), d("2026-12-31")), "2026-09-15")
	// Out of the requested window → nothing.
	assertDates(t, Dates(rule, d("2026-10-01"), d("2026-12-31")))
	// tanggal_akhir_ulang is ignored for a single-date schedule.
	rule.TanggalAkhirUlang = dp("2026-09-01")
	assertDates(t, Dates(rule, d("2026-09-15"), d("2026-09-15")), "2026-09-15")
}

func TestDatesHarian(t *testing.T) {
	rule := Rule{Pola: domain.PolaHarian, Interval: 1, TanggalMulai: d("2026-09-01")}
	assertDates(t, Dates(rule, d("2026-09-01"), d("2026-09-04")),
		"2026-09-01", "2026-09-02", "2026-09-03", "2026-09-04")

	// Interval 3, window starting after tanggal_mulai keeps the phase anchored
	// on tanggal_mulai (01, 04, 07, 10 …).
	rule.Interval = 3
	assertDates(t, Dates(rule, d("2026-09-05"), d("2026-09-12")), "2026-09-07", "2026-09-10")

	// tanggal_akhir_ulang caps the window.
	rule.TanggalAkhirUlang = dp("2026-09-08")
	assertDates(t, Dates(rule, d("2026-09-01"), d("2026-12-31")), "2026-09-01", "2026-09-04", "2026-09-07")
}

func TestDatesMingguan(t *testing.T) {
	// 2026-09-01 is a Tuesday. Mon(1)/Wed(3)/Sat(6) every week.
	rule := Rule{Pola: domain.PolaMingguan, Interval: 1, TanggalMulai: d("2026-09-01"), HariUlang: []int64{1, 3, 6}}
	assertDates(t, Dates(rule, d("2026-09-01"), d("2026-09-14")),
		"2026-09-02", "2026-09-05", "2026-09-07", "2026-09-09", "2026-09-12", "2026-09-14")

	// Every 2nd week, Monday-start weeks anchored on the week of 2026-09-01
	// (Mon 08-31). Weeks: 08-31, [09-07 skipped], 09-14, [09-21 skipped], 09-28.
	rule.Interval = 2
	rule.HariUlang = []int64{3} // Wednesday
	assertDates(t, Dates(rule, d("2026-09-01"), d("2026-09-30")), "2026-09-02", "2026-09-16", "2026-09-30")

	// Empty hari_ulang falls back to tanggal_mulai's weekday (Tuesday).
	rule.Interval = 1
	rule.HariUlang = nil
	assertDates(t, Dates(rule, d("2026-09-01"), d("2026-09-15")), "2026-09-01", "2026-09-08", "2026-09-15")

	// Sunday is 0, matching time.Weekday.
	rule.HariUlang = []int64{0}
	assertDates(t, Dates(rule, d("2026-09-01"), d("2026-09-14")), "2026-09-06", "2026-09-13")
}

func TestDatesBulanan(t *testing.T) {
	rule := Rule{Pola: domain.PolaBulanan, Interval: 1, TanggalMulai: d("2026-01-31")}
	// Feb/Apr/Jun have no 31st → skipped, not clamped.
	assertDates(t, Dates(rule, d("2026-01-01"), d("2026-06-30")), "2026-01-31", "2026-03-31", "2026-05-31")

	rule = Rule{Pola: domain.PolaBulanan, Interval: 3, TanggalMulai: d("2026-09-10")}
	assertDates(t, Dates(rule, d("2026-09-01"), d("2027-06-30")),
		"2026-09-10", "2026-12-10", "2027-03-10", "2027-06-10")
}

func TestDatesKustom(t *testing.T) {
	// Every 2 days, but only weekdays Mon–Fri. 2026-09-01 Tue, 03 Thu, 05 Sat(x),
	// 07 Mon, 09 Wed, 11 Fri, 13 Sun(x), 15 Tue.
	rule := Rule{Pola: domain.PolaKustom, Interval: 2, TanggalMulai: d("2026-09-01"), HariUlang: []int64{1, 2, 3, 4, 5}}
	assertDates(t, Dates(rule, d("2026-09-01"), d("2026-09-15")),
		"2026-09-01", "2026-09-03", "2026-09-07", "2026-09-09", "2026-09-11", "2026-09-15")
}

func TestDatesWindowAndCap(t *testing.T) {
	rule := Rule{Pola: domain.PolaHarian, Interval: 1, TanggalMulai: d("2026-09-01")}
	// dari before tanggal_mulai is clamped; sampai before dari → empty.
	assertDates(t, Dates(rule, d("2026-08-01"), d("2026-09-02")), "2026-09-01", "2026-09-02")
	assertDates(t, Dates(rule, d("2026-09-10"), d("2026-09-05")))

	// The cap stops iteration at maxSesiPerGenerate+1 so the caller can detect
	// overflow without materialising a decade of rows.
	got := Dates(rule, d("2026-01-01"), d("2036-01-01"))
	if len(got) != maxSesiPerGenerate+1 {
		t.Fatalf("expected %d dates at the cap, got %d", maxSesiPerGenerate+1, len(got))
	}
}

func TestParseClock(t *testing.T) {
	c, err := ParseClock("07:00")
	if err != nil || c != (Clock{7, 0, 0}) || c.String() != "07:00:00" {
		t.Fatalf("07:00 → %v, %v", c, err)
	}
	c, err = ParseClock("23:59:59")
	if err != nil || c.Seconds() != 86399 {
		t.Fatalf("23:59:59 → %v, %v", c, err)
	}
	for _, bad := range []string{"7:00", "24:00", "07:60", "07:00:60", "0700", "", "07:00:00:00"} {
		if _, err := ParseClock(bad); err == nil {
			t.Fatalf("%q should be rejected", bad)
		}
	}
}

func TestClockAtUsesScheduleZone(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Jakarta") // UTC+7, no DST
	if err != nil {
		t.Fatal(err)
	}
	got := Clock{7, 0, 0}.At(d("2026-09-06"), loc).UTC()
	want := time.Date(2026, 9, 6, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("07:00 WIB on 2026-09-06 = %s, want %s", got, want)
	}

	// Asia/Jayapura is UTC+9: same wall clock, different instant.
	jayapura, _ := time.LoadLocation("Asia/Jayapura")
	got = Clock{7, 0, 0}.At(d("2026-09-06"), jayapura).UTC()
	want = time.Date(2026, 9, 5, 22, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("07:00 WIT on 2026-09-06 = %s, want %s", got, want)
	}
}

func TestNormalizeHariUlang(t *testing.T) {
	got := normalizeHariUlang([]int{6, 1, 1, 3, 9, -1})
	want := domain.IntArray{1, 3, 6}
	if len(got) != len(want) {
		t.Fatalf("got %v want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("got %v want %v", got, want)
		}
	}
	if normalizeHariUlang(nil) != nil {
		t.Fatal("empty input must stay nil (SQL NULL)")
	}
}
