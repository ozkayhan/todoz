package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"todoz/internal/model"
)

// QueryOptions holds all parsed filter/sort/output flags for the load command.
type QueryOptions struct {
	AfterDate   string
	BeforeDate  string
	DaysBack    int // 0 = unset
	Status      string
	Overdue     bool
	Lists       []string
	Search      string
	NoTrash     bool
	TrashOnly   bool
	SortBy      string
	SortReverse bool
	GroupBy     string
	Fields      []string
	OutputFormat string
	Pretty      bool
	Summary     bool
	Count       bool
	IntervalDays int
}

// QueryResult is the output of ApplyQuery.
type QueryResult struct {
	Tasks   []model.Task
	Groups  map[string][]model.Task
	Summary *QuerySummary
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
	if opts.DaysBack > 0 {
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
	default:
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
		default:
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

func subtractDays(today string, n int) string {
	t, err := time.Parse("2006-01-02", today)
	if err != nil {
		return today
	}
	return t.AddDate(0, 0, -n).Format("2006-01-02")
}

func containsInsensitive(s, sub string) bool {
	if sub == "" {
		return true
	}
	return indexInsensitive(s, sub) >= 0
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
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		b[i] = c
	}
	return string(b)
}

// ParseQueryOptions converts a flags map into QueryOptions.
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
