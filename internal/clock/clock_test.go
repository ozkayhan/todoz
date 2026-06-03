package clock

import (
	"regexp"
	"testing"
	"time"
)

// Format must always emit microsecond precision and the fixed +03:00 offset.
func TestFormatHasMicrosecondsAndIstanbulOffset(t *testing.T) {
	in := time.Date(2026, 6, 3, 10, 30, 45, 123456000, Zone())
	got := Format(in)
	want := "2026-06-03T10:30:45.123456+03:00"
	if got != want {
		t.Fatalf("Format = %q, want %q", got, want)
	}
}

// Now must be expressed in the Istanbul zone regardless of machine zone.
func TestNowUsesIstanbulOffset(t *testing.T) {
	got := Format(Now())
	re := regexp.MustCompile(`\+03:00$`)
	if !re.MatchString(got) {
		t.Fatalf("Now formatted = %q, want suffix +03:00", got)
	}
}
