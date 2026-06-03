package events

import (
	"todoz/internal/model"
	"todoz/internal/state"
)

// Apply applies an event to the state, updating it in place.
func Apply(s *state.State, e Event) {
	switch e.Type {
	case TypeListCreated:
		s.Lists[e.ListID] = model.List{
			ID:        e.ListID,
			Name:      e.ListName,
			CreatedAt: e.At,
		}
	case TypeTaskCreated:
		s.Tasks[e.TaskID] = model.Task{
			ID:          e.TaskID,
			Title:       e.Title,
			Description: e.Description,
			ListID:      e.ListID,
			Date:        e.Date,
			Status:      model.StatusPending,
			CreatedAt:   e.At,
			UpdatedAt:   e.At,
		}
	}
}
