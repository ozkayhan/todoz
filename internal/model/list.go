package model

// List is a named container of tasks. Deleting a list soft-deletes the list and
// all of its tasks (they move to trash; nothing is physically removed).
type List struct {
	// ID is the unique identifier, e.g. "list-ab12...".
	ID string `json:"id"`
	// Name is the human-readable list name.
	Name string `json:"name"`
	// CreatedAt is the canonical todoz creation timestamp.
	CreatedAt string `json:"createdAt"`
	// IsDeleted marks the list as soft-deleted.
	IsDeleted bool `json:"isDeleted"`
	// DeletedAt is set when the list is soft-deleted (else empty).
	DeletedAt string `json:"deletedAt,omitempty"`
}
