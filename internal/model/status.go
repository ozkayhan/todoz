// Package model defines the pure data structures of todoz (Task, List) and the
// pure functions over them. It performs no I/O and has no dependencies on disk,
// time sources, or the event log, which keeps it trivially testable.
package model

// Status is the lifecycle state of a task. A task is only ever pending or
// completed. "Overdue" is NOT a status — it is a computed display flag (see
// ComputeOverdue) derived from the task's date.
type Status string

const (
	// StatusPending means the task has not been completed.
	StatusPending Status = "pending"
	// StatusCompleted means the user marked the task done.
	StatusCompleted Status = "completed"
)

// Valid reports whether s is a recognized status.
func (s Status) Valid() bool {
	return s == StatusPending || s == StatusCompleted
}
