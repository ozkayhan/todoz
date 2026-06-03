package events

import "testing"

func TestEventTypesDistinct(t *testing.T) {
	all := []string{
		TypeListCreated, TypeListUpdated, TypeListDeleted,
		TypeTaskCreated, TypeTaskUpdated, TypeTaskCompleted,
		TypeTaskDeleted, TypeTaskPermanentlyDeleted, TypeTaskRestored,
	}
	seen := map[string]bool{}
	for _, ty := range all {
		if ty == "" {
			t.Fatal("empty event type constant")
		}
		if seen[ty] {
			t.Fatalf("duplicate event type %q", ty)
		}
		seen[ty] = true
	}
}
