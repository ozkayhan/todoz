package store

import (
	"path/filepath"
	"testing"
	"todoz/internal/events"
)

func TestReadAllMissingFileIsEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "events.jsonl")
	got, err := ReadAll(path)
	if err != nil {
		t.Fatalf("ReadAll missing: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("want empty, got %d", len(got))
	}
}

func TestAppendThenReadAll(t *testing.T) {
	path := filepath.Join(t.TempDir(), "events.jsonl")
	e1 := events.Event{Type: events.TypeListCreated, At: "T1", ListID: "l1", ListName: "A"}
	e2 := events.Event{Type: events.TypeTaskCreated, At: "T2", TaskID: "t1", ListID: "l1", Title: "x", Date: "2026-06-05"}
	if err := AppendEvent(path, e1); err != nil {
		t.Fatalf("append e1: %v", err)
	}
	if err := AppendEvent(path, e2); err != nil {
		t.Fatalf("append e2: %v", err)
	}
	got, err := ReadAll(path)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 events, got %d", len(got))
	}
	if got[0].Type != e1.Type || got[0].At != e1.At || got[0].ListID != e1.ListID || got[0].ListName != e1.ListName {
		t.Fatalf("round trip mismatch for e1: %+v", got[0])
	}
	if got[1].Type != e2.Type || got[1].At != e2.At || got[1].TaskID != e2.TaskID || got[1].ListID != e2.ListID || got[1].Title != e2.Title || got[1].Date != e2.Date {
		t.Fatalf("round trip mismatch for e2: %+v", got[1])
	}
}
