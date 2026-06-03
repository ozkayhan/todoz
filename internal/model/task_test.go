package model

import (
	"encoding/json"
	"testing"
)

// Task must serialize with the exact JSON field names the contract promises.
func TestTaskJSONFields(t *testing.T) {
	task := Task{
		ID:        "task-1",
		Title:     "Buy milk",
		ListID:    "list-1",
		Date:      "2026-06-05",
		Status:    StatusPending,
		CreatedAt: "2026-06-03T10:00:00.000000+03:00",
		UpdatedAt: "2026-06-03T10:00:00.000000+03:00",
	}
	b, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, key := range []string{"id", "title", "listId", "date", "status", "isOverdue", "createdAt", "updatedAt"} {
		if _, ok := m[key]; !ok {
			t.Fatalf("missing JSON key %q in %s", key, b)
		}
	}
}

// List must serialize with the exact JSON field names.
func TestListJSONFields(t *testing.T) {
	l := List{ID: "list-1", Name: "Groceries", CreatedAt: "2026-06-03T10:00:00.000000+03:00"}
	b, _ := json.Marshal(l)
	var m map[string]any
	_ = json.Unmarshal(b, &m)
	for _, key := range []string{"id", "name", "createdAt"} {
		if _, ok := m[key]; !ok {
			t.Fatalf("missing JSON key %q in %s", key, b)
		}
	}
}
