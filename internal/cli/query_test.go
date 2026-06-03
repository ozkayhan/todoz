package cli

import (
	"testing"
	"todoz/internal/model"
)

func TestApplyQuery_NoFilters_ReturnsAllActive(t *testing.T) {
	tasks := []model.Task{
		{ID: "t1", Title: "A", Date: "2026-06-01", Status: model.StatusPending, ListID: "l1"},
		{ID: "t2", Title: "B", Date: "2026-06-02", Status: model.StatusCompleted, ListID: "l1"},
	}
	lists := []model.List{{ID: "l1", Name: "Work"}}
	result := ApplyQuery(tasks, lists, QueryOptions{}, "2026-06-03")
	if len(result.Tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(result.Tasks))
	}
}
