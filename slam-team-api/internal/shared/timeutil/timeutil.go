// Package timeutil holds the project's time conventions. Stored and computed
// times are always timestamptz in UTC; an IANA zone is applied only at the
// presentation edge, from the row's own `timezone` column — never from the
// server or browser zone.
package timeutil

import "time"

// DefaultZone is the fallback IANA zone (users.timezone default).
const DefaultZone = "Asia/Jakarta"

// Location resolves an IANA zone name, falling back to UTC when the name is
// empty or unknown. Resolution relies on the tzdata embedded by cmd/api and
// cmd/slamctl (import _ "time/tzdata"), so it works on scratch/alpine images.
func Location(iana string) *time.Location {
	if iana == "" {
		return time.UTC
	}
	loc, err := time.LoadLocation(iana)
	if err != nil {
		return time.UTC
	}
	return loc
}

// InZone renders t in the given IANA zone (UTC if the zone is unknown). The
// instant is unchanged — only its wall-clock presentation.
func InZone(t time.Time, iana string) time.Time {
	return t.In(Location(iana))
}

// NowUTC is the single source of "now" for service code.
func NowUTC() time.Time { return time.Now().UTC() }
