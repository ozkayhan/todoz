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

func makeTasks() []model.Task {
	return []model.Task{
		{ID: "t1", Title: "Alpha", Date: "2026-05-25", Status: model.StatusPending, ListID: "l1", Description: "urgent"},
		{ID: "t2", Title: "Beta", Date: "2026-05-30", Status: model.StatusCompleted, ListID: "l1"},
		{ID: "t3", Title: "Gamma", Date: "2026-06-02", Status: model.StatusPending, ListID: "l2"},
		{ID: "t4", Title: "Delta", Date: "2026-06-05", Status: model.StatusPending, ListID: "l2"},
	}
}

func TestApplyQuery_DaysBack(t *testing.T) {
	tasks := makeTasks()
	opts := QueryOptions{DaysBack: 7}
	result := ApplyQuery(tasks, nil, opts, "2026-06-03")
	if len(result.Tasks) != 3 {
		t.Fatalf("want 3, got %d: %v", len(result.Tasks), result.Tasks)
	}
}

func TestApplyQuery_AfterDateBeforeDate(t *testing.T) {
	tasks := makeTasks()
	opts := QueryOptions{AfterDate: "2026-05-28", BeforeDate: "2026-06-03"}
	result := ApplyQuery(tasks, nil, opts, "2026-06-03")
	if len(result.Tasks) != 2 {
		t.Fatalf("want 2, got %d", len(result.Tasks))
	}
}

func TestApplyQuery_StatusPending(t *testing.T) {
	tasks := makeTasks()
	opts := QueryOptions{Status: "pending"}
	result := ApplyQuery(tasks, nil, opts, "2026-06-03")
	for _, tk := range result.Tasks {
		if tk.Status != model.StatusPending {
			t.Fatalf("non-pending task in result: %v", tk)
		}
	}
	if len(result.Tasks) != 3 {
		t.Fatalf("want 3 pending, got %d", len(result.Tasks))
	}
}

func TestApplyQuery_Overdue(t *testing.T) {
	tasks := makeTasks()
	opts := QueryOptions{Overdue: true}
	result := ApplyQuery(tasks, nil, opts, "2026-06-03")
	if len(result.Tasks) != 2 {
		t.Fatalf("want 2 overdue, got %d", len(result.Tasks))
	}
}

func TestApplyQuery_FilterLists(t *testing.T) {
	tasks := makeTasks()
	opts := QueryOptions{Lists: []string{"l1"}}
	result := ApplyQuery(tasks, nil, opts, "2026-06-03")
	if len(result.Tasks) != 2 {
		t.Fatalf("want 2, got %d", len(result.Tasks))
	}
}

func TestApplyQuery_Search(t *testing.T) {
	tasks := makeTasks()
	opts := QueryOptions{Search: "URG"}
	result := ApplyQuery(tasks, nil, opts, "2026-06-03")
	if len(result.Tasks) != 1 || result.Tasks[0].ID != "t1" {
		t.Fatalf("want t1 only, got %v", result.Tasks)
	}
}

func TestApplyQuery_MultipleFilters(t *testing.T) {
	tasks := makeTasks()
	opts := QueryOptions{Status: "pending", Lists: []string{"l2"}, AfterDate: "2026-06-03"}
	result := ApplyQuery(tasks, nil, opts, "2026-06-03")
	if len(result.Tasks) != 1 || result.Tasks[0].ID != "t4" {
		t.Fatalf("want t4 only, got %v", result.Tasks)
	}
}

func TestApplyQuery_SortByTitle(t *testing.T) {
	tasks := makeTasks()
	opts := QueryOptions{SortBy: "title"}
	result := ApplyQuery(tasks, nil, opts, "2026-06-03")
	want := []string{"Alpha", "Beta", "Delta", "Gamma"}
	for i, tk := range result.Tasks {
		if tk.Title != want[i] {
			t.Fatalf("pos %d: want %s got %s", i, want[i], tk.Title)
		}
	}
}

func TestApplyQuery_SortByDateReverse(t *testing.T) {
	tasks := makeTasks()
	opts := QueryOptions{SortBy: "date", SortReverse: true}
	result := ApplyQuery(tasks, nil, opts, "2026-06-03")
	if result.Tasks[0].ID != "t4" {
		t.Fatalf("want t4 first, got %s", result.Tasks[0].ID)
	}
}

func TestApplyQuery_GroupByList(t *testing.T) {
	tasks := makeTasks()
	lists := []model.List{
		{ID: "l1", Name: "Work"},
		{ID: "l2", Name: "Personal"},
	}
	opts := QueryOptions{GroupBy: "list"}
	result := ApplyQuery(tasks, lists, opts, "2026-06-03")
	if len(result.Groups["Work"]) != 2 {
		t.Fatalf("want 2 in Work, got %d", len(result.Groups["Work"]))
	}
	if len(result.Groups["Personal"]) != 2 {
		t.Fatalf("want 2 in Personal, got %d", len(result.Groups["Personal"]))
	}
}

func TestApplyQuery_GroupByDate(t *testing.T) {
	tasks := makeTasks()
	opts := QueryOptions{GroupBy: "date"}
	result := ApplyQuery(tasks, nil, opts, "2026-06-03")
	if len(result.Groups["2026-06-05"]) != 1 {
		t.Fatalf("want 1 in 2026-06-05, got %d", len(result.Groups["2026-06-05"]))
	}
}

func TestApplyQuery_GroupByStatus(t *testing.T) {
	tasks := makeTasks()
	opts := QueryOptions{GroupBy: "status"}
	result := ApplyQuery(tasks, nil, opts, "2026-06-03")
	if len(result.Groups["pending"]) != 3 {
		t.Fatalf("want 3 pending, got %d", len(result.Groups["pending"]))
	}
	if len(result.Groups["completed"]) != 1 {
		t.Fatalf("want 1 completed, got %d", len(result.Groups["completed"]))
	}
}

func TestApplyQuery_Summary(t *testing.T) {
	tasks := makeTasks()
	opts := QueryOptions{Summary: true}
	result := ApplyQuery(tasks, nil, opts, "2026-06-03")
	if result.Summary == nil {
		t.Fatal("want summary, got nil")
	}
	if result.Summary.Total != 4 {
		t.Fatalf("want total=4, got %d", result.Summary.Total)
	}
	if result.Summary.Pending != 3 {
		t.Fatalf("want pending=3, got %d", result.Summary.Pending)
	}
	if result.Summary.Completed != 1 {
		t.Fatalf("want completed=1, got %d", result.Summary.Completed)
	}
	if result.Summary.Overdue != 2 {
		t.Fatalf("want overdue=2, got %d", result.Summary.Overdue)
	}
}

func TestApplyQuery_Count(t *testing.T) {
	tasks := makeTasks()
	opts := QueryOptions{Count: true}
	result := ApplyQuery(tasks, nil, opts, "2026-06-03")
	if result.Summary == nil {
		t.Fatal("want summary for count, got nil")
	}
	if result.Summary.Total != 4 {
		t.Fatalf("want total=4, got %d", result.Summary.Total)
	}
}

func TestParseQueryOptions_DaysBack(t *testing.T) {
	flags := map[string]string{"days-back": "7"}
	opts, err := ParseQueryOptions(flags)
	if err != nil {
		t.Fatal(err)
	}
	if opts.DaysBack != 7 {
		t.Fatalf("want DaysBack=7, got %d", opts.DaysBack)
	}
}

func TestParseQueryOptions_Lists(t *testing.T) {
	flags := map[string]string{"lists": "l1,l2,l3"}
	opts, err := ParseQueryOptions(flags)
	if err != nil {
		t.Fatal(err)
	}
	if len(opts.Lists) != 3 || opts.Lists[1] != "l2" {
		t.Fatalf("want [l1,l2,l3], got %v", opts.Lists)
	}
}

func TestParseQueryOptions_Fields(t *testing.T) {
	flags := map[string]string{"fields": "id,title,date"}
	opts, err := ParseQueryOptions(flags)
	if err != nil {
		t.Fatal(err)
	}
	if len(opts.Fields) != 3 {
		t.Fatalf("want 3 fields, got %v", opts.Fields)
	}
}

func TestParseQueryOptions_InvalidDaysBack(t *testing.T) {
	flags := map[string]string{"days-back": "notanumber"}
	_, err := ParseQueryOptions(flags)
	if err == nil {
		t.Fatal("want error for non-numeric days-back")
	}
}
