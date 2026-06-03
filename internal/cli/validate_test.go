package cli

import "testing"

func TestValidDate(t *testing.T) {
	if !ValidDate("2026-06-05") {
		t.Fatal("2026-06-05 should be valid")
	}
	for _, bad := range []string{"2026-6-5", "06-05-2026", "2026/06/05", "", "not-a-date", "2026-13-01"} {
		if ValidDate(bad) {
			t.Fatalf("%q should be invalid", bad)
		}
	}
}
