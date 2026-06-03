package store

import (
	"os"
	"todoz/internal/events"
	"todoz/internal/model"
	"todoz/internal/state"
)

// Compact rewrites the event log, replacing the entire history with a snapshot
// of the current state. This reduces file size and load time by truncating
// historical operations that no longer affect the final state.
func (s *Store) Compact() error {
	lock, err := Acquire(s.cfg.LockPath(), lockTimeoutMS)
	if err != nil {
		return err
	}
	defer lock.Release()

	evs, err := ReadAll(s.cfg.EventsPath())
	if err != nil {
		return err
	}

	st := state.New()
	for _, e := range evs {
		events.Apply(&st, e)
	}

	snapshot := snapshotEvents(st)
	tmp := s.cfg.EventsPath() + ".tmp"
	if err := writeEvents(tmp, snapshot); err != nil {
		return err
	}

	_ = os.Rename(s.cfg.EventsPath(), s.cfg.BackupPath())
	return os.Rename(tmp, s.cfg.EventsPath())
}

// snapshotEvents converts the current state into a minimal set of events that
// reconstruct it (without intermediate history).
func snapshotEvents(st state.State) []events.Event {
	var out []events.Event

	for _, l := range st.Lists {
		out = append(out, events.Event{
			Type: events.TypeListCreated, At: l.CreatedAt,
			ListID: l.ID, ListName: l.Name,
		})
		if l.IsDeleted {
			out = append(out, events.Event{
				Type: events.TypeListDeleted, At: l.DeletedAt,
				ListID: l.ID,
			})
		}
	}

	for _, tk := range st.Tasks {
		out = append(out, events.Event{
			Type:        events.TypeTaskCreated,
			At:          tk.CreatedAt,
			TaskID:      tk.ID,
			ListID:      tk.ListID,
			Title:       tk.Title,
			Description: tk.Description,
			Date:        tk.Date,
		})
		if tk.Status == model.StatusCompleted {
			out = append(out, events.Event{
				Type:   events.TypeTaskCompleted,
				At:     tk.CompletedAt,
				TaskID: tk.ID,
			})
		}
		if tk.IsHiddenTrash {
			out = append(out, events.Event{
				Type:   events.TypeTaskPermanentlyDeleted,
				At:     tk.UpdatedAt,
				TaskID: tk.ID,
			})
		} else if tk.IsDeleted {
			out = append(out, events.Event{
				Type:   events.TypeTaskDeleted,
				At:     tk.DeletedAt,
				TaskID: tk.ID,
			})
		}
	}

	return out
}

// writeEvents writes a list of events to a file, creating it if necessary.
func writeEvents(path string, evs []events.Event) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, e := range evs {
		line, err := events.Encode(e)
		if err != nil {
			return err
		}
		if _, err := f.WriteString(line + "\n"); err != nil {
			return err
		}
	}

	return f.Sync()
}
