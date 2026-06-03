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
	case TypeTaskDeleted:
		if t, ok := s.Tasks[e.TaskID]; ok {
			t.IsDeleted = true
			t.DeletedAt = e.At
			t.UpdatedAt = e.At
			s.Tasks[e.TaskID] = t
		}
	case TypeTaskPermanentlyDeleted:
		if t, ok := s.Tasks[e.TaskID]; ok {
			t.IsDeleted = true
			t.IsHiddenTrash = true
			t.UpdatedAt = e.At
			s.Tasks[e.TaskID] = t
		}
	case TypeTaskRestored:
		if t, ok := s.Tasks[e.TaskID]; ok {
			t.IsDeleted = false
			t.IsHiddenTrash = false
			t.DeletedAt = ""
			t.UpdatedAt = e.At
			s.Tasks[e.TaskID] = t
		}
	case TypeListUpdated:
		if l, ok := s.Lists[e.ListID]; ok {
			if v, has := e.Updates["name"]; has {
				l.Name = v
			}
			s.Lists[e.ListID] = l
		}
	case TypeListDeleted:
		if l, ok := s.Lists[e.ListID]; ok {
			l.IsDeleted = true
			s.Lists[e.ListID] = l
		}
		// Cascade: soft-delete all tasks in this list
		for id, t := range s.Tasks {
			if t.ListID == e.ListID && !t.IsDeleted {
				t.IsDeleted = true
				t.DeletedAt = e.At
				t.UpdatedAt = e.At
				s.Tasks[id] = t
			}
		}
	}
}
