package store

import (
	"todoz/internal/config"
	"todoz/internal/events"
	"todoz/internal/state"
)

const lockTimeoutMS = 5000

// Store provides locked access to the event log for appends and full-replay
// loading of the application state.
type Store struct {
	cfg config.Config
}

// New creates a new Store for the given configuration.
func New(c config.Config) *Store {
	return &Store{cfg: c}
}

// Load reads all events from the log and replays them to construct the
// current application state.
func (s *Store) Load() (state.State, error) {
	evs, err := ReadAll(s.cfg.EventsPath())
	if err != nil {
		return state.State{}, err
	}

	st := state.New()
	for _, e := range evs {
		events.Apply(&st, e)
	}

	return st, nil
}

// Append atomically appends an event to the log while holding an exclusive lock.
func (s *Store) Append(e events.Event) error {
	lock, err := Acquire(s.cfg.LockPath(), lockTimeoutMS)
	if err != nil {
		return err
	}
	defer lock.Release()

	return AppendEvent(s.cfg.EventsPath(), e)
}
