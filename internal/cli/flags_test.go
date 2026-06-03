package cli

import "testing"

func TestParseFlags(t *testing.T) {
	args := []string{"--title", "Buy milk", "--date", "2026-06-05", "--list", "l1"}
	f := ParseFlags(args)
	if f["title"] != "Buy milk" || f["date"] != "2026-06-05" || f["list"] != "l1" {
		t.Fatalf("flags parsed wrong: %+v", f)
	}
}

func TestParseBooleanFlag(t *testing.T) {
	f := ParseFlags([]string{"--permanently"})
	if f["permanently"] != "true" {
		t.Fatalf("bool flag wrong: %+v", f)
	}
}

func TestParsePositional(t *testing.T) {
	f := ParseFlags([]string{"task-1"})
	if f["_"] != "task-1" {
		t.Fatalf("positional wrong: %+v", f)
	}
}
