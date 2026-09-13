package service

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"slam-team-api/internal/modules/core/jadwal/domain"
)

// maxSesiPerGenerate bounds one generate-sesi call. A year of daily sessions
// is 365 rows; anything past this is almost certainly a wrong date range, and
// the call is idempotent so a client can always ask for the next range.
const maxSesiPerGenerate = 1000

// Rule is the recurrence input the generator reads, decoupled from the
// entity so the algorithm can be unit-tested without a database.
//
// Semantics per pola:
//   - tidak_berulang: exactly one date, tanggal_mulai (hari_ulang ignored).
//   - harian:   tanggal_mulai, then every Interval days.
//   - mingguan: the weekdays in HariUlang (0 = Minggu … 6 = Sabtu; empty →
//     tanggal_mulai's weekday), in every Interval-th week counted from the
//     Monday-start week containing tanggal_mulai.
//   - bulanan:  tanggal_mulai's day-of-month every Interval months; months
//     without that day (31 in a 30-day month, 29–31 in February) are skipped.
//   - kustom:   every Interval days like harian, additionally filtered to the
//     weekdays in HariUlang when it is non-empty. (The DBML leaves kustom
//     undefined; this is the most general rule the stored columns express.)
//
// TanggalAkhirUlang, when set, caps recurring patterns. The generator's own
// window [dari, sampai] is intersected with that.
type Rule struct {
	Pola              domain.PolaUlang
	HariUlang         []int64
	Interval          int
	TanggalMulai      time.Time
	TanggalAkhirUlang *time.Time
}

// Dates returns every date the rule produces inside [dari, sampai]
// (inclusive), ascending, as UTC-midnight time.Time values. It never returns
// more than maxSesiPerGenerate+1 entries — callers check the length.
func Dates(rule Rule, dari, sampai time.Time) []time.Time {
	interval := rule.Interval
	if interval < 1 {
		interval = 1
	}
	mulai := dateOnly(rule.TanggalMulai)
	lo := dateOnly(dari)
	hi := dateOnly(sampai)
	if mulai.After(lo) {
		lo = mulai
	}
	if rule.TanggalAkhirUlang != nil && rule.Pola != domain.PolaTidakBerulang {
		if akhir := dateOnly(*rule.TanggalAkhirUlang); akhir.Before(hi) {
			hi = akhir
		}
	}

	var out []time.Time
	add := func(d time.Time) bool {
		out = append(out, d)
		return len(out) <= maxSesiPerGenerate
	}

	switch rule.Pola {
	case domain.PolaHarian, domain.PolaKustom:
		var days map[int]bool // harian never filters by weekday
		if rule.Pola == domain.PolaKustom {
			days = weekdaySet(rule.HariUlang)
		}
		for d := mulai; !d.After(hi); d = d.AddDate(0, 0, interval) {
			if d.Before(lo) {
				continue
			}
			if days != nil && !days[int(d.Weekday())] {
				continue
			}
			if !add(d) {
				return out
			}
		}

	case domain.PolaMingguan:
		days := weekdaySet(rule.HariUlang)
		if days == nil {
			days = map[int]bool{int(mulai.Weekday()): true}
		}
		anchor := mondayOf(mulai)
		for d := mulai; !d.After(hi); d = d.AddDate(0, 0, 1) {
			if d.Before(lo) {
				continue
			}
			weeks := int(mondayOf(d).Sub(anchor).Hours() / 24 / 7)
			if weeks%interval != 0 || !days[int(d.Weekday())] {
				continue
			}
			if !add(d) {
				return out
			}
		}

	case domain.PolaBulanan:
		day := mulai.Day()
		for k := 0; ; k += interval {
			first := time.Date(mulai.Year(), mulai.Month()+time.Month(k), 1, 0, 0, 0, 0, time.UTC)
			if first.After(hi) {
				break
			}
			d := first.AddDate(0, 0, day-1)
			if d.Month() != first.Month() { // month too short for this day
				continue
			}
			if d.Before(lo) {
				continue
			}
			if d.After(hi) {
				break
			}
			if !add(d) {
				return out
			}
		}

	default: // tidak_berulang (and anything unknown) → the single start date
		if !mulai.Before(lo) && !mulai.After(hi) {
			add(mulai)
		}
	}
	return out
}

// dateOnly strips the clock and pins the date to UTC midnight so date maths
// never crosses a DST boundary.
func dateOnly(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// mondayOf returns the Monday that starts the week containing d.
func mondayOf(d time.Time) time.Time {
	back := (int(d.Weekday()) + 6) % 7 // Mon=0 … Sun=6
	return d.AddDate(0, 0, -back)
}

// weekdaySet turns hari_ulang into a lookup; nil when empty.
func weekdaySet(hari []int64) map[int]bool {
	if len(hari) == 0 {
		return nil
	}
	set := make(map[int]bool, len(hari))
	for _, h := range hari {
		set[int(h)] = true
	}
	return set
}

// normalizeHariUlang dedupes and sorts the weekday list; nil when empty.
func normalizeHariUlang(in []int) domain.IntArray {
	if len(in) == 0 {
		return nil
	}
	seen := make(map[int]bool, len(in))
	out := make(domain.IntArray, 0, len(in))
	for _, h := range in {
		if h < 0 || h > 6 || seen[h] {
			continue
		}
		seen[h] = true
		out = append(out, int64(h))
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// Clock is a time-of-day without a date.
type Clock struct{ H, M, S int }

// ParseClock accepts "HH:MM" or "HH:MM:SS".
func ParseClock(s string) (Clock, error) {
	parts := strings.Split(strings.TrimSpace(s), ":")
	if len(parts) != 2 && len(parts) != 3 {
		return Clock{}, fmt.Errorf("format jam harus HH:MM atau HH:MM:SS")
	}
	nums := make([]int, 3)
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || len(p) != 2 {
			return Clock{}, fmt.Errorf("format jam harus HH:MM atau HH:MM:SS")
		}
		nums[i] = n
	}
	c := Clock{H: nums[0], M: nums[1], S: nums[2]}
	if c.H > 23 || c.M > 59 || c.S > 59 {
		return Clock{}, fmt.Errorf("jam di luar rentang 00:00:00–23:59:59")
	}
	return c, nil
}

// String renders the PG `time` text form.
func (c Clock) String() string { return fmt.Sprintf("%02d:%02d:%02d", c.H, c.M, c.S) }

// Seconds since midnight, for ordering comparisons.
func (c Clock) Seconds() int { return c.H*3600 + c.M*60 + c.S }

// At combines the clock with a date in loc — the only place local wall time
// becomes an instant. Callers convert with .UTC().
func (c Clock) At(date time.Time, loc *time.Location) time.Time {
	y, m, d := date.Date()
	return time.Date(y, m, d, c.H, c.M, c.S, 0, loc)
}

// parseDate reads a YYYY-MM-DD string as a UTC-midnight date.
func parseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}

// fmtDate renders a date column as YYYY-MM-DD.
func fmtDate(t time.Time) string { return t.Format("2006-01-02") }
