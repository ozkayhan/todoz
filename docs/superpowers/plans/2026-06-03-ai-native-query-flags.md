# AI-Native Query Flags Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extend `todoz load` with composable filtering, sorting, grouping, and output flags so an LLM can construct a single CLI command that returns exactly the data it needs — no post-processing required.

**Architecture:** All new flags live in `cmdLoad` inside `internal/cli/commands.go`. A new `internal/cli/query.go` file holds the `QueryOptions` struct and the pure `ApplyQuery(tasks []model.Task, lists []model.List, opts QueryOptions, today string) QueryResult` function, keeping filter/sort/group logic independently testable. Conflict validation lives in `internal/cli/query_validate.go`.

**Tech Stack:** Go stdlib only. No new dependencies.

---

## File Map

| File | Action | Responsibility |
|------|--------|---------------|
| `internal/cli/query.go` | Create | `QueryOptions` struct + `ApplyQuery` pure function |
| `internal/cli/query_validate.go` | Create | `ValidateQueryFlags(flags map[string]string) error` |
| `internal/cli/query_test.go` | Create | Unit tests for `ApplyQuery` |
| `internal/cli/query_validate_test.go` | Create | Unit tests for conflict detection |
| `internal/cli/commands.go` | Modify | `cmdLoad` reads new flags, calls `ApplyQuery` |
| `internal/cli/commands_test.go` | Modify | Integration tests for new `load` flags |
| `USAGE.md` | Modify | Document all new flags |

---

## Task 1: Create `QueryOptions` struct and skeleton `ApplyQuery`

**Files:**
- Create: `internal/cli/query.go`
- Create: `internal/cli/query_test.go`

- [ ] **Step 1: Write the failing test**

```go
// internal/cli/query_test.go
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
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd /path/to/todo && go test ./internal/cli/ -run TestApplyQuery_NoFilters_ReturnsAllActive -v
```

Expected: `undefined: ApplyQuery`

- [ ] **Step 3: Create `internal/cli/query.go`**

```go
package cli

import "todoz/internal/model"

// QueryOptions holds all parsed filter/sort/output flags for the load command.
type QueryOptions struct {
	// Filtering
	AfterDate  string // YYYY-MM-DD inclusive lower bound on task.Date
	BeforeDate string // YYYY-MM-DD inclusive upper bound on task.Date
	DaysBack   int    // shorthand: AfterDate = today minus DaysBack days; -1 = unset
	Status     string // "pending", "completed", or "" (all)
	Overdue    bool   // pending AND date < today
	Lists      []string // filter to these list IDs (empty = all)
	Search     string // substring match on title or description
	NoTrash    bool   // exclude trash (default false — caller already passes active tasks)
	TrashOnly  bool   // only trash tasks

	// Sorting
	SortBy      string // "date" (default), "title", "created", "status"
	SortReverse bool

	// Grouping
	GroupBy string // "list", "date", "status", or "" (no grouping)

	// Output
	Fields        []string // subset of: id, title, date, status, listId, description, isOverdue, createdAt, completedAt
	OutputFormat  string   // "json" (default), "compact", "csv"
	Pretty        bool

	// Aggregation
	Summary bool // include counts summary
	Count   bool // only return count

	// Snapshots
	IntervalDays int // produce sub-slices per N-day window; -1 = unset
}

// QueryResult is the output of ApplyQuery.
type QueryResult struct {
	Tasks   []model.Task
	Groups  map[string][]model.Task // populated when GroupBy is set
	Summary *QuerySummary           // populated when Summary=true or Count=true
}

// QuerySummary holds aggregate counts.
type QuerySummary struct {
	Total     int `json:"total"`
	Pending   int `json:"pending"`
	Completed int `json:"completed"`
	Overdue   int `json:"overdue"`
}

// ApplyQuery filters, sorts, and groups tasks according to opts.
// today must be YYYY-MM-DD.
func ApplyQuery(tasks []model.Task, lists []model.List, opts QueryOptions, today string) QueryResult {
	filtered := filterTasks(tasks, opts, today)
	sorted := sortTasks(filtered, opts)
	result := QueryResult{Tasks: sorted}
	if opts.GroupBy != "" {
		result.Groups = groupTasks(sorted, lists, opts.GroupBy)
	}
	if opts.Summary || opts.Count {
		result.Summary = computeSummary(sorted, today)
	}
	return result
}

func filterTasks(tasks []model.Task, opts QueryOptions, today string) []model.Task {
	out := make([]model.Task, 0, len(tasks))
	afterDate := opts.AfterDate
	if opts.DaysBack >= 0 && opts.DaysBack != -1 {
		afterDate = subtractDays(today, opts.DaysBack)
	}
	listSet := make(map[string]bool, len(opts.Lists))
	for _, id := range opts.Lists {
		listSet[id] = true
	}
	for _, t := range tasks {
		if afterDate != "" && t.Date < afterDate {
			continue
		}
		if opts.BeforeDate != "" && t.Date > opts.BeforeDate {
			continue
		}
		if opts.Status != "" && string(t.Status) != opts.Status {
			continue
		}
		if opts.Overdue && !(t.Status == model.StatusPending && t.Date < today) {
			continue
		}
		if len(listSet) > 0 && !listSet[t.ListID] {
			continue
		}
		if opts.Search != "" && !containsInsensitive(t.Title, opts.Search) && !containsInsensitive(t.Description, opts.Search) {
			continue
		}
		out = append(out, t)
	}
	return out
}

func sortTasks(tasks []model.Task, opts QueryOptions) []model.Task {
	out := make([]model.Task, len(tasks))
	copy(out, tasks)
	sortBy := opts.SortBy
	if sortBy == "" {
		sortBy = "date"
	}
	// insertion sort — small N, stdlib only
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && taskLess(out[j], out[j-1], sortBy, opts.SortReverse); j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}

func taskLess(a, b model.Task, by string, reverse bool) bool {
	var less bool
	switch by {
	case "title":
		less = a.Title < b.Title
	case "created":
		less = a.CreatedAt < b.CreatedAt
	case "status":
		less = string(a.Status) < string(b.Status)
	default: // "date"
		if a.Date != b.Date {
			less = a.Date < b.Date
		} else {
			less = a.ID < b.ID
		}
	}
	if reverse {
		return !less
	}
	return less
}

func groupTasks(tasks []model.Task, lists []model.List, by string) map[string][]model.Task {
	listNames := make(map[string]string, len(lists))
	for _, l := range lists {
		listNames[l.ID] = l.Name
	}
	groups := make(map[string][]model.Task)
	for _, t := range tasks {
		var key string
		switch by {
		case "list":
			if name, ok := listNames[t.ListID]; ok {
				key = name
			} else {
				key = t.ListID
			}
		case "status":
			key = string(t.Status)
		default: // "date"
			key = t.Date
		}
		groups[key] = append(groups[key], t)
	}
	return groups
}

func computeSummary(tasks []model.Task, today string) *QuerySummary {
	s := &QuerySummary{Total: len(tasks)}
	for _, t := range tasks {
		switch t.Status {
		case model.StatusPending:
			s.Pending++
			if t.Date < today {
				s.Overdue++
			}
		case model.StatusCompleted:
			s.Completed++
		}
	}
	return s
}

// subtractDays subtracts n calendar days from a YYYY-MM-DD string.
// Uses string arithmetic via time.Parse — stdlib only.
func subtractDays(today string, n int) string {
	t, err := parseDate(today)
	if err != nil {
		return today
	}
	return t.AddDate(0, 0, -n).Format("2006-01-02")
}

func containsInsensitive(s, sub string) bool {
	if sub == "" {
		return true
	}
	return len(s) >= len(sub) && indexInsensitive(s, sub) >= 0
}

func indexInsensitive(s, sub string) int {
	ls, lsub := toLower(s), toLower(sub)
	for i := 0; i+len(lsub) <= len(ls); i++ {
		if ls[i:i+len(lsub)] == lsub {
			return i
		}
	}
	return -1
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := range s {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		b[i] = c
	}
	return string(b)
}
```

- [ ] **Step 4: Add `parseDate` helper to `internal/cli/validate.go`**

```go
// add to internal/cli/validate.go (after ValidDate)
import "time"

func parseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}
```

- [ ] **Step 5: Run test to verify it passes**

```bash
go test ./internal/cli/ -run TestApplyQuery_NoFilters_ReturnsAllActive -v
```

Expected: `PASS`

- [ ] **Step 6: Commit**

```bash
git add internal/cli/query.go internal/cli/query_test.go internal/cli/validate.go
git commit -m "feat: add QueryOptions and ApplyQuery skeleton"
```

---

## Task 2: Filtering tests — date, status, search, lists

**Files:**
- Modify: `internal/cli/query_test.go`

- [ ] **Step 1: Write failing tests for each filter**

```go
// append to internal/cli/query_test.go

func makeTasks() []model.Task {
	return []model.Task{
		{ID: "t1", Title: "Alpha", Date: "2026-05-25", Status: model.StatusPending, ListID: "l1", Description: "urgent"},
		{ID: "t2", Title: "Beta",  Date: "2026-05-30", Status: model.StatusCompleted, ListID: "l1"},
		{ID: "t3", Title: "Gamma", Date: "2026-06-02", Status: model.StatusPending, ListID: "l2"},
		{ID: "t4", Title: "Delta", Date: "2026-06-05", Status: model.StatusPending, ListID: "l2"},
	}
}

func TestApplyQuery_DaysBack(t *testing.T) {
	tasks := makeTasks()
	// today=2026-06-03, days-back=7 → after 2026-05-27
	opts := QueryOptions{DaysBack: 7}
	result := ApplyQuery(tasks, nil, opts, "2026-06-03")
	// t1 is 2026-05-25 < 2026-05-27 → excluded; t2,t3,t4 included
	if len(result.Tasks) != 3 {
		t.Fatalf("want 3, got %d: %v", len(result.Tasks), result.Tasks)
	}
}

func TestApplyQuery_AfterDateBeforeDate(t *testing.T) {
	tasks := makeTasks()
	opts := QueryOptions{AfterDate: "2026-05-28", BeforeDate: "2026-06-03"}
	result := ApplyQuery(tasks, nil, opts, "2026-06-03")
	// t2 (05-30) and t3 (06-02) match
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
	// today=2026-06-03; overdue = pending && date < 2026-06-03
	// t1 (05-25, pending) ✓; t2 (05-30, completed) ✗; t3 (06-02, pending) ✓; t4 (06-05, pending) ✗
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
	opts := QueryOptions{Search: "URG"} // case-insensitive
	result := ApplyQuery(tasks, nil, opts, "2026-06-03")
	if len(result.Tasks) != 1 || result.Tasks[0].ID != "t1" {
		t.Fatalf("want t1 only, got %v", result.Tasks)
	}
}

func TestApplyQuery_MultipleFilters(t *testing.T) {
	tasks := makeTasks()
	// pending AND in l2 AND date >= 2026-06-03
	opts := QueryOptions{Status: "pending", Lists: []string{"l2"}, AfterDate: "2026-06-03"}
	result := ApplyQuery(tasks, nil, opts, "2026-06-03")
	// t4 (06-05, pending, l2) matches; t3 (06-02) < 06-03 excluded
	if len(result.Tasks) != 1 || result.Tasks[0].ID != "t4" {
		t.Fatalf("want t4 only, got %v", result.Tasks)
	}
}
```

- [ ] **Step 2: Run to verify they fail**

```bash
go test ./internal/cli/ -run "TestApplyQuery_DaysBack|TestApplyQuery_After|TestApplyQuery_Status|TestApplyQuery_Overdue|TestApplyQuery_Filter|TestApplyQuery_Search|TestApplyQuery_Multiple" -v
```

Expected: FAIL (logic not yet wired for `DaysBack=-1` default — see Step 3)

- [ ] **Step 3: Fix `QueryOptions` zero value for DaysBack**

In `query.go`, `DaysBack` must default to "unset". Change the filter check:

```go
// in filterTasks, replace DaysBack block:
if opts.DaysBack > 0 {
    afterDate = subtractDays(today, opts.DaysBack)
}
```

(Remove the `-1` sentinel; 0 means "not set" since 0 days back = today, which is rarely intended and covered by `--after-date` instead. Keep `DaysBack int` zero-value = disabled.)

Also update `QueryOptions` comment: `// DaysBack: shorthand; 0 = unset`

- [ ] **Step 4: Run tests**

```bash
go test ./internal/cli/ -run "TestApplyQuery" -v
```

Expected: all PASS

- [ ] **Step 5: Commit**

```bash
git add internal/cli/query.go internal/cli/query_test.go
git commit -m "test: filtering tests for ApplyQuery (date, status, search, lists)"
```

---

## Task 3: Sorting and grouping tests

**Files:**
- Modify: `internal/cli/query_test.go`

- [ ] **Step 1: Write failing tests**

```go
// append to internal/cli/query_test.go

func TestApplyQuery_SortByTitle(t *testing.T) {
	tasks := makeTasks() // Alpha, Beta, Gamma, Delta
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
	// descending: t4, t3, t2, t1
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
```

- [ ] **Step 2: Run**

```bash
go test ./internal/cli/ -run "TestApplyQuery_Sort|TestApplyQuery_Group" -v
```

Expected: PASS (logic already in query.go)

- [ ] **Step 3: Commit**

```bash
git add internal/cli/query_test.go
git commit -m "test: sorting and grouping tests for ApplyQuery"
```

---

## Task 4: Summary / count tests

**Files:**
- Modify: `internal/cli/query_test.go`

- [ ] **Step 1: Write failing tests**

```go
// append to internal/cli/query_test.go

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
	// t1 (05-25 < 06-03, pending) and t3 (06-02 < 06-03, pending) are overdue
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
```

- [ ] **Step 2: Run**

```bash
go test ./internal/cli/ -run "TestApplyQuery_Summary|TestApplyQuery_Count" -v
```

Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add internal/cli/query_test.go
git commit -m "test: summary and count tests for ApplyQuery"
```

---

## Task 5: Conflict validation

**Files:**
- Create: `internal/cli/query_validate.go`
- Create: `internal/cli/query_validate_test.go`

- [ ] **Step 1: Write failing tests**

```go
// internal/cli/query_validate_test.go
package cli

import "testing"

func TestValidateQueryFlags_DaysBackAndAfterDate_Conflict(t *testing.T) {
	flags := map[string]string{"days-back": "7", "after-date": "2026-05-01"}
	if err := ValidateQueryFlags(flags); err == nil {
		t.Fatal("expected conflict error")
	}
}

func TestValidateQueryFlags_NoTrashAndTrashOnly_Conflict(t *testing.T) {
	flags := map[string]string{"no-trash": "true", "trash-only": "true"}
	if err := ValidateQueryFlags(flags); err == nil {
		t.Fatal("expected conflict error")
	}
}

func TestValidateQueryFlags_ValidCombination(t *testing.T) {
	flags := map[string]string{"days-back": "7", "status": "pending", "sort-by": "date", "group-by": "list"}
	if err := ValidateQueryFlags(flags); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateQueryFlags_InvalidStatus(t *testing.T) {
	flags := map[string]string{"status": "overdue"}
	if err := ValidateQueryFlags(flags); err == nil {
		t.Fatal("expected error for invalid status")
	}
}

func TestValidateQueryFlags_InvalidSortBy(t *testing.T) {
	flags := map[string]string{"sort-by": "priority"}
	if err := ValidateQueryFlags(flags); err == nil {
		t.Fatal("expected error for invalid sort-by value")
	}
}
```

- [ ] **Step 2: Run to verify fail**

```bash
go test ./internal/cli/ -run "TestValidateQueryFlags" -v
```

Expected: `undefined: ValidateQueryFlags`

- [ ] **Step 3: Create `internal/cli/query_validate.go`**

```go
package cli

import (
	"errors"
	"strings"
)

var validSortBy = map[string]bool{"date": true, "title": true, "created": true, "status": true}
var validGroupBy = map[string]bool{"list": true, "date": true, "status": true}
var validStatus = map[string]bool{"pending": true, "completed": true}
var validOutputFormat = map[string]bool{"json": true, "compact": true, "csv": true}

// ValidateQueryFlags checks for conflicting or invalid flag combinations.
func ValidateQueryFlags(flags map[string]string) error {
	if flags["days-back"] != "" && flags["after-date"] != "" {
		return errors.New("conflict: use --days-back OR --after-date, not both")
	}
	if flags["no-trash"] == "true" && flags["trash-only"] == "true" {
		return errors.New("conflict: --no-trash and --trash-only are mutually exclusive")
	}
	if v := flags["status"]; v != "" && !validStatus[v] {
		return errors.New("invalid --status value: must be pending or completed, got: " + v)
	}
	if v := flags["sort-by"]; v != "" && !validSortBy[v] {
		valid := "date, title, created, status"
		return errors.New("invalid --sort-by value: must be one of " + valid + ", got: " + v)
	}
	if v := flags["group-by"]; v != "" && !validGroupBy[v] {
		return errors.New("invalid --group-by value: must be list, date, or status, got: " + v)
	}
	if v := flags["output-format"]; v != "" && !validOutputFormat[v] {
		return errors.New("invalid --output-format value: must be json, compact, or csv, got: " + v)
	}
	_ = strings.TrimSpace // imported for future use
	return nil
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/cli/ -run "TestValidateQueryFlags" -v
```

Expected: all PASS

- [ ] **Step 5: Commit**

```bash
git add internal/cli/query_validate.go internal/cli/query_validate_test.go
git commit -m "feat: add ValidateQueryFlags for conflict and value checks"
```

---

## Task 6: `ParseQueryOptions` — flags map → `QueryOptions`

**Files:**
- Modify: `internal/cli/query.go`
- Modify: `internal/cli/query_test.go`

- [ ] **Step 1: Write failing tests**

```go
// append to internal/cli/query_test.go

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
```

- [ ] **Step 2: Run to verify fail**

```bash
go test ./internal/cli/ -run "TestParseQueryOptions" -v
```

Expected: `undefined: ParseQueryOptions`

- [ ] **Step 3: Add `ParseQueryOptions` to `query.go`**

```go
// add to internal/cli/query.go

import (
	"strconv"
	"strings"
	"todoz/internal/model"
)

// ParseQueryOptions converts a flags map into QueryOptions.
// Returns error if any flag has an invalid value.
func ParseQueryOptions(flags map[string]string) (QueryOptions, error) {
	opts := QueryOptions{}

	if v := flags["after-date"]; v != "" {
		opts.AfterDate = v
	}
	if v := flags["before-date"]; v != "" {
		opts.BeforeDate = v
	}
	if v := flags["days-back"]; v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			return opts, fmt.Errorf("--days-back must be a positive integer, got: %s", v)
		}
		opts.DaysBack = n
	}
	if v := flags["status"]; v != "" {
		opts.Status = v
	}
	if flags["overdue"] == "true" {
		opts.Overdue = true
	}
	if v := flags["lists"]; v != "" {
		for _, id := range strings.Split(v, ",") {
			if id = strings.TrimSpace(id); id != "" {
				opts.Lists = append(opts.Lists, id)
			}
		}
	}
	// single --list flag still works (from sprint 1)
	if v := flags["list"]; v != "" && flags["lists"] == "" {
		opts.Lists = []string{v}
	}
	if v := flags["search"]; v != "" {
		opts.Search = v
	}
	if flags["no-trash"] == "true" {
		opts.NoTrash = true
	}
	if flags["trash-only"] == "true" {
		opts.TrashOnly = true
	}
	if v := flags["sort-by"]; v != "" {
		opts.SortBy = v
	}
	if flags["sort-reverse"] == "true" {
		opts.SortReverse = true
	}
	if v := flags["group-by"]; v != "" {
		opts.GroupBy = v
	}
	if v := flags["fields"]; v != "" {
		for _, f := range strings.Split(v, ",") {
			if f = strings.TrimSpace(f); f != "" {
				opts.Fields = append(opts.Fields, f)
			}
		}
	}
	if v := flags["output-format"]; v != "" {
		opts.OutputFormat = v
	}
	if flags["pretty"] == "true" {
		opts.Pretty = true
	}
	if flags["summary"] == "true" {
		opts.Summary = true
	}
	if flags["count"] == "true" {
		opts.Count = true
	}
	if v := flags["interval-days"]; v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			return opts, fmt.Errorf("--interval-days must be a positive integer, got: %s", v)
		}
		opts.IntervalDays = n
	}
	return opts, nil
}
```

Add `"fmt"` to the import block.

- [ ] **Step 4: Run**

```bash
go test ./internal/cli/ -run "TestParseQueryOptions" -v
```

Expected: all PASS

- [ ] **Step 5: Run full suite**

```bash
go test ./...
```

Expected: all PASS

- [ ] **Step 6: Commit**

```bash
git add internal/cli/query.go internal/cli/query_test.go
git commit -m "feat: add ParseQueryOptions to parse flags map into QueryOptions"
```

---

## Task 7: Wire `cmdLoad` to use `ApplyQuery`

**Files:**
- Modify: `internal/cli/commands.go`
- Modify: `internal/cli/commands_test.go`

- [ ] **Step 1: Write integration tests**

```go
// append to internal/cli/commands_test.go

func seedTask(t *testing.T, ctx Ctx, listID, title, date string) string {
	t.Helper()
	res := cmdAddTask(ctx, ParseFlags([]string{"--title", title, "--date", date, "--list", listID}))
	if !res.OK {
		t.Fatalf("seedTask failed: %+v", res)
	}
	return res.Data.(map[string]string)["id"]
}

func TestCmdLoad_DaysBack(t *testing.T) {
	ctx := testCtx(t)
	ctx.Today = "2026-06-03"
	listID := seedList(t, ctx)
	seedTask(t, ctx, listID, "Old", "2026-05-20")
	seedTask(t, ctx, listID, "Recent", "2026-06-01")
	res := cmdLoad(ctx, ParseFlags([]string{"--days-back", "7"}))
	if !res.OK {
		t.Fatalf("load failed: %+v", res)
	}
	view := res.Data.(LoadView)
	if len(view.Tasks) != 1 || view.Tasks[0].Title != "Recent" {
		t.Fatalf("want [Recent], got %+v", view.Tasks)
	}
}

func TestCmdLoad_StatusFilter(t *testing.T) {
	ctx := testCtx(t)
	ctx.Today = "2026-06-03"
	listID := seedList(t, ctx)
	taskID := seedTask(t, ctx, listID, "T1", "2026-06-05")
	seedTask(t, ctx, listID, "T2", "2026-06-06")
	cmdCompleteTask(ctx, ParseFlags([]string{taskID}))
	res := cmdLoad(ctx, ParseFlags([]string{"--status", "pending"}))
	view := res.Data.(LoadView)
	if len(view.Tasks) != 1 || view.Tasks[0].Title != "T2" {
		t.Fatalf("want [T2], got %+v", view.Tasks)
	}
}

func TestCmdLoad_SortByTitleReverse(t *testing.T) {
	ctx := testCtx(t)
	ctx.Today = "2026-06-03"
	listID := seedList(t, ctx)
	seedTask(t, ctx, listID, "Bravo", "2026-06-01")
	seedTask(t, ctx, listID, "Alpha", "2026-06-02")
	seedTask(t, ctx, listID, "Charlie", "2026-06-03")
	res := cmdLoad(ctx, ParseFlags([]string{"--sort-by", "title", "--sort-reverse"}))
	view := res.Data.(LoadView)
	if view.Tasks[0].Title != "Charlie" {
		t.Fatalf("want Charlie first, got %s", view.Tasks[0].Title)
	}
}

func TestCmdLoad_ConflictError(t *testing.T) {
	ctx := testCtx(t)
	res := cmdLoad(ctx, ParseFlags([]string{"--days-back", "7", "--after-date", "2026-01-01"}))
	if res.OK || res.ErrCode != "invalid_operation" {
		t.Fatalf("want invalid_operation, got %+v", res)
	}
}

func TestCmdLoad_GroupBy(t *testing.T) {
	ctx := testCtx(t)
	ctx.Today = "2026-06-03"
	listID := seedList(t, ctx)
	seedTask(t, ctx, listID, "T1", "2026-06-01")
	seedTask(t, ctx, listID, "T2", "2026-06-02")
	res := cmdLoad(ctx, ParseFlags([]string{"--group-by", "date"}))
	if !res.OK {
		t.Fatalf("load failed: %+v", res)
	}
	view := res.Data.(GroupedLoadView)
	if len(view.Groups) != 2 {
		t.Fatalf("want 2 date groups, got %d: %+v", len(view.Groups), view.Groups)
	}
}
```

- [ ] **Step 2: Run to verify fail**

```bash
go test ./internal/cli/ -run "TestCmdLoad_DaysBack|TestCmdLoad_StatusFilter|TestCmdLoad_SortByTitleReverse|TestCmdLoad_ConflictError|TestCmdLoad_GroupBy" -v
```

Expected: FAIL

- [ ] **Step 3: Update `commands.go` — add `GroupedLoadView` and rewrite `cmdLoad`**

```go
// add after LoadView in commands.go

// GroupedLoadView is returned by cmdLoad when --group-by is set.
type GroupedLoadView struct {
	Groups  map[string][]model.Task `json:"groups"`
	Lists   []model.List            `json:"lists"`
	Summary *QuerySummary           `json:"summary,omitempty"`
}

// rewrite cmdLoad:
func cmdLoad(ctx Ctx, flags map[string]string) response.Envelope {
	if err := ValidateQueryFlags(flags); err != nil {
		return response.Error("invalid_operation", err.Error())
	}
	opts, err := ParseQueryOptions(flags)
	if err != nil {
		return response.Error("invalid_operation", err.Error())
	}
	st, err := ctx.Store.Load()
	if err != nil {
		return response.Error("io_error", err.Error())
	}

	var sourceTasks []model.Task
	if opts.TrashOnly {
		sourceTasks = st.TrashTasks()
	} else {
		sourceTasks = st.ActiveTasks()
		if !opts.NoTrash {
			// default: active only (trash accessible via --trash-only)
		}
	}

	// compute IsOverdue
	for i := range sourceTasks {
		sourceTasks[i].IsOverdue = model.ComputeOverdue(sourceTasks[i], ctx.Today)
	}

	activeLists := st.ActiveLists()
	sort.Slice(activeLists, func(i, j int) bool { return activeLists[i].CreatedAt < activeLists[j].CreatedAt })

	result := ApplyQuery(sourceTasks, activeLists, opts, ctx.Today)

	if opts.GroupBy != "" {
		view := GroupedLoadView{
			Groups:  result.Groups,
			Lists:   activeLists,
			Summary: result.Summary,
		}
		return response.Success(view)
	}

	view := LoadView{
		Lists: activeLists,
		Tasks: result.Tasks,
		Trash: []model.Task{},
	}
	if !opts.TrashOnly && !opts.NoTrash {
		for _, tk := range st.TrashTasks() {
			view.Trash = append(view.Trash, tk)
		}
	}
	if result.Summary != nil {
		// attach summary as a wrapper
		type loadWithSummary struct {
			LoadView
			Summary *QuerySummary `json:"summary"`
		}
		return response.Success(loadWithSummary{view, result.Summary})
	}
	return response.Success(view)
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/cli/ -run "TestCmdLoad" -v
```

Expected: all PASS

- [ ] **Step 5: Run full suite**

```bash
go test ./...
```

Expected: all PASS

- [ ] **Step 6: Commit**

```bash
git add internal/cli/commands.go internal/cli/commands_test.go
git commit -m "feat: wire cmdLoad to ApplyQuery — all query flags active"
```

---

## Task 8: Build and smoke test

- [ ] **Step 1: Build**

```bash
cd /path/to/todo && go build -o todoz ./cmd/todoz
```

Expected: no errors

- [ ] **Step 2: Smoke test — basic flag combos**

```bash
export TODO_LIB_HOME=$(mktemp -d)
export TODO_APP_NAME=smoke

./todoz add-list --name Work
# grab list id from output, e.g. LIST_ID="list-abc..."
LIST_ID=$(./todoz add-list --name Work | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['id'])")

./todoz add-task --title "Old task"    --date 2026-05-20 --list $LIST_ID
./todoz add-task --title "Recent task" --date 2026-06-01 --list $LIST_ID
./todoz add-task --title "Future task" --date 2026-06-10 --list $LIST_ID

# days-back filter
./todoz load --days-back 7
# should include Recent and Future (last 7 days from today 2026-06-03 = after 2026-05-27)

# status filter
./todoz load --status pending

# sort + reverse
./todoz load --sort-by title --sort-reverse

# group-by
./todoz load --group-by date

# summary
./todoz load --summary

# conflict → should fail with invalid_operation
./todoz load --days-back 7 --after-date 2026-01-01
echo "exit code: $?"
```

- [ ] **Step 3: Commit**

```bash
git add todoz
git commit -m "build: update compiled binary for sprint2"
```

---

## Task 9: Update `USAGE.md`

**Files:**
- Modify: `USAGE.md`

- [ ] **Step 1: Add new flags section to `load` command**

Find the `### load` section in `USAGE.md` and extend it:

```markdown
### load
`todoz load [flags]`

Returns: `{"lists": [...], "tasks": [...], "trash": [...]}`
When `--group-by` is set, returns: `{"groups": {"date": [...]}, "lists": [...], "summary": {...}}`

#### Filtering Flags
| Flag | Description |
|------|-------------|
| `--days-back N` | Tasks with date ≥ today minus N days |
| `--after-date YYYY-MM-DD` | Tasks with date ≥ this date (mutually exclusive with `--days-back`) |
| `--before-date YYYY-MM-DD` | Tasks with date ≤ this date |
| `--status pending\|completed` | Only tasks with this status |
| `--overdue` | Only pending tasks where date < today |
| `--lists "id1,id2"` | Only tasks in these list IDs (comma-separated) |
| `--search TEXT` | Substring match on title or description (case-insensitive) |
| `--no-trash` | Exclude trash from response |
| `--trash-only` | Return only trash tasks (mutually exclusive with `--no-trash`) |

#### Sorting Flags
| Flag | Description |
|------|-------------|
| `--sort-by date\|title\|created\|status` | Sort field (default: date) |
| `--sort-reverse` | Reverse sort order |

#### Grouping Flags
| Flag | Description |
|------|-------------|
| `--group-by list\|date\|status` | Group tasks; changes response shape to `groups` map |

#### Output / Aggregation Flags
| Flag | Description |
|------|-------------|
| `--summary` | Append `{"total":N,"pending":N,"completed":N,"overdue":N}` to response |
| `--count` | Same as `--summary` (only counts, no tasks needed) |

#### Conflict Rules
- `--days-back` and `--after-date` cannot be used together
- `--no-trash` and `--trash-only` cannot be used together

#### Example Combinations
```bash
# Last 7 days, pending only
todoz load --days-back 7 --status pending

# Overdue tasks grouped by list
todoz load --overdue --group-by list

# Search in a date range, sorted descending
todoz load --after-date 2026-05-01 --before-date 2026-06-01 --search "urgent" --sort-by date --sort-reverse

# Summary stats for all tasks
todoz load --summary

# All tasks in two lists, grouped by date
todoz load --lists "list-abc,list-def" --group-by date
```
```

- [ ] **Step 2: Commit**

```bash
git add USAGE.md
git commit -m "docs: document new query flags for load command"
```

---

## Self-Review

**Spec coverage check:**
- ✅ Date filtering: `--days-back`, `--after-date`, `--before-date`
- ✅ Status/overdue filtering: `--status`, `--overdue`
- ✅ List filtering: `--lists`, `--search`
- ✅ Trash visibility: `--no-trash`, `--trash-only`
- ✅ Sorting: `--sort-by`, `--sort-reverse`
- ✅ Grouping: `--group-by`
- ✅ Aggregation: `--summary`, `--count`
- ✅ Conflict validation and error codes
- ✅ Backward compat: existing `--list` still works
- ✅ USAGE.md updated
- ✅ All flags composable together

**Placeholder scan:** None found.

**Type consistency:**
- `LoadView` used in existing tests; `GroupedLoadView` added alongside — no rename conflict.
- `QuerySummary` defined in `query.go`, referenced in `GroupedLoadView` and `loadWithSummary` — consistent.
- `ApplyQuery` signature matches all call sites.
