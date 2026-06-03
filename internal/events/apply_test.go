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

func TestApplyTaskUpdated(t *testing.T) {
	s := state.New()
	Apply(&s, Event{Type: TypeTaskCreated, At: "T1", TaskID: "task-1", ListID: "l1", Title: "Old", Date: "2026-06-05"})
	Apply(&s, Event{Type: TypeTaskUpdated, At: "T2", TaskID: "task-1", Updates: map[string]string{"title": "New", "description": "d", "date": "2026-06-06"}})
	task := s.Tasks["task-1"]
	if task.Title != "New" || task.Description != "d" || task.Date != "2026-06-06" {
		t.Fatalf("update not applied: %+v", task)
	}
	if task.UpdatedAt != "T2" {
		t.Fatalf("UpdatedAt=%q, want T2", task.UpdatedAt)
	}
}

func TestApplyTaskCompleted(t *testing.T) {
	s := state.New()
	Apply(&s, Event{Type: TypeTaskCreated, At: "T1", TaskID: "task-1", ListID: "l1", Title: "x", Date: "2026-06-05"})
	Apply(&s, Event{Type: TypeTaskCompleted, At: "T2", TaskID: "task-1"})
	task := s.Tasks["task-1"]
	if task.Status != "completed" || task.CompletedAt != "T2" {
		t.Fatalf("complete not applied: %+v", task)
	}
}
