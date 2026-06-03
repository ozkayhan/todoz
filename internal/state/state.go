package state

import "todoz/internal/model"

// State is the in-memory container of all lists and tasks.
type State struct {
	Lists map[string]model.List
	Tasks map[string]model.Task
}

// New creates an empty State.
func New() State {
	return State{
		Lists: make(map[string]model.List),
		Tasks: make(map[string]model.Task),
	}
}

// ActiveTasks returns tasks that are not deleted and not hidden trash.
func (s State) ActiveTasks() []model.Task {
	out := make([]model.Task, 0, len(s.Tasks))
	for _, t := range s.Tasks {
		if !t.IsDeleted && !t.IsHiddenTrash {
			out = append(out, t)
		}
	}
	return out
}

// TrashTasks returns tasks that are deleted but not hidden trash.
func (s State) TrashTasks() []model.Task {
	out := make([]model.Task, 0, len(s.Tasks))
	for _, t := range s.Tasks {
		if t.IsDeleted && !t.IsHiddenTrash {
			out = append(out, t)
		}
	}
	return out
}

// ActiveLists returns lists that are not deleted.
func (s State) ActiveLists() []model.List {
	out := make([]model.List, 0, len(s.Lists))
	for _, l := range s.Lists {
		if !l.IsDeleted {
			out = append(out, l)
		}
	}
	return out
}
