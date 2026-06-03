package model

// Task is a single todo item. A task belongs to exactly one list and is tied to
// a single calendar date (no time-of-day). Tasks are never physically deleted;
// deletion sets IsDeleted (soft, visible in trash) and then IsHiddenTrash
// (permanent, hidden from apps).
type Task struct {
	// ID is the unique identifier, e.g. "task-ab12...".
	ID string `json:"id"`
	// Title is the short task label (required).
	Title string `json:"title"`
	// Description is optional free text.
	Description string `json:"description"`
	// ListID is the owning list (required).
	ListID string `json:"listId"`
	// Date is the assigned day in YYYY-MM-DD form (no time component).
	Date string `json:"date"`
	// Status is pending or completed.
	Status Status `json:"status"`
	// IsOverdue is a computed flag: pending AND Date < today. Never stored in
	// the event log; populated by ComputeOverdue at read time.
	IsOverdue bool `json:"isOverdue"`
	// IsDeleted marks the task as soft-deleted (in trash, still visible to apps).
	IsDeleted bool `json:"isDeleted"`
	// IsHiddenTrash marks the task as permanently deleted (hidden from apps).
	IsHiddenTrash bool `json:"isHiddenTrash"`
	// CreatedAt/UpdatedAt are canonical todoz timestamps.
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	// CompletedAt/DeletedAt are set when those transitions occur (else empty).
	CompletedAt string `json:"completedAt,omitempty"`
	DeletedAt   string `json:"deletedAt,omitempty"`
}
