package model

// ComputeOverdue reports whether a task should display as overdue relative to
// the given today date (YYYY-MM-DD). A task is overdue when it is still pending
// and its date is strictly before today. Because the date format is zero-padded
// ISO (YYYY-MM-DD), lexical string comparison is equivalent to date comparison.
func ComputeOverdue(t Task, today string) bool {
	return t.Status == StatusPending && t.Date < today
}
