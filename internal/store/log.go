package store

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"todoz/internal/events"
)

// ReadAll reads all events from the event log file. If the file does not exist,
// returns an empty slice (no error).
func ReadAll(path string) ([]events.Event, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []events.Event{}, nil
		}
		return nil, err
	}
	defer f.Close()

	var out []events.Event
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		e, err := events.Decode(line)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}

	if err := sc.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

// AppendEvent appends a single event to the event log file, creating the file
// and its parent directories if necessary.
func AppendEvent(path string, e events.Event) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	line, err := events.Encode(e)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString(line + "\n"); err != nil {
		return err
	}

	return f.Sync()
}
