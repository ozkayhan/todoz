package events

const (
	TypeListCreated = "list_created"
	TypeListUpdated = "list_updated"
	TypeListDeleted = "list_deleted"

	TypeTaskCreated            = "task_created"
	TypeTaskUpdated            = "task_updated"
	TypeTaskCompleted          = "task_completed"
	TypeTaskDeleted            = "task_deleted"
	TypeTaskPermanentlyDeleted = "task_permanently_deleted"
	TypeTaskRestored           = "task_restored"
)

// Event represents a single mutation in the application state.
type Event struct {
	Type string `json:"type"`
	At   string `json:"at"`

	// List fields
	ListID   string `json:"listId,omitempty"`
	ListName string `json:"name,omitempty"`

	// Task fields
	TaskID      string `json:"taskId,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Date        string `json:"date,omitempty"`

	// Updates is a map of field names to new values (used in update events)
	Updates map[string]string `json:"updates,omitempty"`
}
