package store

import (
	"testing"
	"todoz/internal/events"
)

func TestCompactPreservesState(t *testing.T) {
	s := newTestStore(t)
	_ = s.Append(events.Event{Type: events.TypeListCreated, At: "T1", ListID: "l1", ListName: "A"})
	_ = s.Append(events.Event{Type: events.TypeTaskCreated, At: "T2", TaskID: "t1", ListID: "l1", Title: "x", Date: "2026-06-05"})
	_ = s.Append(events.Event{Type: events.TypeTaskCompleted, At: "T3", TaskID: "t1"})
	before, _ := s.Load()
	if err := s.Compact(); err != nil {
		t.Fatalf("compact: %v", err)
	}
	after, _ := s.Load()
	if len(after.Lists) != len(before.Lists) || len(after.Tasks) != len(before.Tasks) {
		t.Fatalf("state changed across compaction: before=%+v after=%+v", before, after)
	}
	if after.Tasks["t1"].Status != "completed" {
		t.Fatalf("task status lost: %+v", after.Tasks["t1"])
	}
}
