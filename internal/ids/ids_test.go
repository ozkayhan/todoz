package ids

import (
	"strings"
	"testing"
)

// New must return a non-empty, prefixed id.
func TestNewHasPrefixAndLength(t *testing.T) {
	got := New("task")
	if !strings.HasPrefix(got, "task-") {
		t.Fatalf("New(\"task\") = %q, want prefix task-", got)
	}
	if len(got) < len("task-")+16 {
		t.Fatalf("New id too short: %q", got)
	}
}

// New must not collide across many calls.
func TestNewIsUnique(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 10000; i++ {
		id := New("task")
		if seen[id] {
			t.Fatalf("duplicate id generated: %q", id)
		}
		seen[id] = true
	}
}
