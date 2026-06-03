package state

import (
	"testing"
	"todoz/internal/model"
)

func TestNewIsEmpty(t *testing.T) {
	s := New()
	if len(s.Lists) != 0 || len(s.Tasks) != 0 {
		t.Fatal("new state should be empty")
	}
}

func TestTaskViews(t *testing.T) {
	s := New()
	s.Tasks["a"] = model.Task{ID: "a"}
	s.Tasks["b"] = model.Task{ID: "b", IsDeleted: true}
	s.Tasks["c"] = model.Task{ID: "c", IsDeleted: true, IsHiddenTrash: true}

	if got := len(s.ActiveTasks()); got != 1 {
		t.Fatalf("ActiveTasks len=%d, want 1", got)
	}
	if got := len(s.TrashTasks()); got != 1 {
		t.Fatalf("TrashTasks len=%d, want 1", got)
	}
}
