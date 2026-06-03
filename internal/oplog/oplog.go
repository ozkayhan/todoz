package oplog

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Entry represents a single operation recorded in the audit log.
type Entry struct {
	Timestamp  string
	Request    string
	OK         bool
	Data       string
	Error      string
	Message    string
	DurationMS int64
}

// Record writes an operation log entry to the log file, creating it and its
// parent directories if necessary.
func Record(path string, e Entry) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	var b strings.Builder
	fmt.Fprintf(&b, "[%s]\n", e.Timestamp)
	fmt.Fprintf(&b, "REQUEST: %s\n", e.Request)
	fmt.Fprintf(&b, "RESPONSE: ok=%t\n", e.OK)

	if e.OK {
		fmt.Fprintf(&b, "DATA: %s\n", e.Data)
	} else {
		fmt.Fprintf(&b, "ERROR: %s\n", e.Error)
		fmt.Fprintf(&b, "MESSAGE: %s\n", e.Message)
	}

	fmt.Fprintf(&b, "DURATION: %dms\n\n", e.DurationMS)

	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(b.String())
	return err
}
