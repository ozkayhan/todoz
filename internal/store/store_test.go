package store

import (
	"testing"
	"todoz/internal/config"
	"todoz/internal/events"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	t.Setenv("TODO_LIB_HOME", t.TempDir())
	t.Setenv("TODO_APP_NAME", "test")
	c, err := config.Load()
	if err != nil {
		t.Fatalf("config: %v", err)
	}
	return New(c)
}

func TestStoreAppendThenLoad(t *testing.T) {
	s := newTestStore(t)
	if err := s.Append(events.Event{Type: events.TypeListCreated, At: "T1", ListID: "l1", ListName: "Groceries"}); err != nil {
		t.Fatalf("append: %v", err)
	}
	st, err := s.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if st.Lists["l1"].Name != "Groceries" {
		t.Fatalf("state missing list: %+v", st.Lists)
	}
}
