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

func TestApplyTaskDeleteRestore(t *testing.T) {
	s := state.New()
	Apply(&s, Event{Type: TypeTaskCreated, At: "T1", TaskID: "task-1", ListID: "l1", Title: "x", Date: "2026-06-05"})
	Apply(&s, Event{Type: TypeTaskDeleted, At: "T2", TaskID: "task-1"})
	if !s.Tasks["task-1"].IsDeleted || s.Tasks["task-1"].DeletedAt != "T2" {
		t.Fatalf("soft delete failed: %+v", s.Tasks["task-1"])
	}
	Apply(&s, Event{Type: TypeTaskRestored, At: "T3", TaskID: "task-1"})
	if s.Tasks["task-1"].IsDeleted {
		t.Fatalf("restore failed: %+v", s.Tasks["task-1"])
	}
	Apply(&s, Event{Type: TypeTaskDeleted, At: "T4", TaskID: "task-1"})
	Apply(&s, Event{Type: TypeTaskPermanentlyDeleted, At: "T5", TaskID: "task-1"})
	if !s.Tasks["task-1"].IsHiddenTrash {
		t.Fatalf("hidden trash failed: %+v", s.Tasks["task-1"])
	}
}

func TestApplyListUpdateDelete(t *testing.T) {
	s := state.New()
	Apply(&s, Event{Type: TypeListCreated, At: "T1", ListID: "l1", ListName: "Old"})
	Apply(&s, Event{Type: TypeListUpdated, At: "T2", ListID: "l1", Updates: map[string]string{"name": "New"}})
	if s.Lists["l1"].Name != "New" {
		t.Fatalf("list update failed: %+v", s.Lists["l1"])
	}
	Apply(&s, Event{Type: TypeTaskCreated, At: "T3", TaskID: "t1", ListID: "l1", Title: "x", Date: "2026-06-05"})
	Apply(&s, Event{Type: TypeListDeleted, At: "T4", ListID: "l1"})
	if !s.Lists["l1"].IsDeleted {
		t.Fatal("list not soft-deleted")
	}
	if !s.Tasks["t1"].IsDeleted {
		t.Fatal("task in deleted list not soft-deleted")
	}
}
