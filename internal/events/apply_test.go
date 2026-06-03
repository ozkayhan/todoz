package events

import (
	"testing"
	"todoz/internal/state"
)

func TestApplyListCreated(t *testing.T) {
	s := state.New()
	Apply(&s, Event{Type: TypeListCreated, At: "T", ListID: "list-1", ListName: "Groceries"})
	l, ok := s.Lists["list-1"]
	if !ok || l.Name != "Groceries" || l.CreatedAt != "T" {
		t.Fatalf("list not created correctly: %+v", l)
	}
}

func TestApplyTaskCreated(t *testing.T) {
	s := state.New()
	Apply(&s, Event{Type: TypeTaskCreated, At: "T", TaskID: "task-1", ListID: "list-1", Title: "Buy milk", Date: "2026-06-05"})
	task, ok := s.Tasks["task-1"]
	if !ok {
		t.Fatal("task not created")
	}
	if task.Title != "Buy milk" || task.Status != "pending" || task.CreatedAt != "T" {
		t.Fatalf("task fields wrong: %+v", task)
	}
}
