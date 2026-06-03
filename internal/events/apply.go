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
	case TypeTaskUpdated:
		if t, ok := s.Tasks[e.TaskID]; ok {
			if v, has := e.Updates["title"]; has {
				t.Title = v
			}
			if v, has := e.Updates["description"]; has {
				t.Description = v
			}
			if v, has := e.Updates["date"]; has {
				t.Date = v
			}
			t.UpdatedAt = e.At
			s.Tasks[e.TaskID] = t
		}
	case TypeTaskCompleted:
		if t, ok := s.Tasks[e.TaskID]; ok {
			t.Status = model.StatusCompleted
			t.CompletedAt = e.At
			t.UpdatedAt = e.At
			s.Tasks[e.TaskID] = t
		}
	}
}
