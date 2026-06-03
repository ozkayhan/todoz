// Package clock is the single source of time for todoz.
//
// Every timestamp in the system (events, operation log, responses) is produced
// here so the format is uniform: RFC3339 with microsecond precision and a fixed
// +03:00 (Istanbul) offset. Istanbul has observed UTC+3 with no daylight saving
// since 2016, so a fixed zone is correct and requires no network/TZ database.
package clock

import "time"

// layout is RFC3339 with 6 fractional digits and a numeric offset.
const layout = "2006-01-02T15:04:05.000000-07:00"

// istanbul is the fixed +03:00 zone used for all timestamps.
var istanbul = time.FixedZone("+03:00", 3*60*60)

// Zone returns the fixed Istanbul (+03:00) location.
func Zone() *time.Location { return istanbul }

// Now returns the current wall-clock time expressed in the Istanbul zone.
func Now() time.Time { return time.Now().In(istanbul) }

// Format renders t using the canonical todoz timestamp layout.
func Format(t time.Time) string { return t.In(istanbul).Format(layout) }
