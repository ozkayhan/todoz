package oplog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRecordWritesBlock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "logs", "app.log")
	err := Record(path, Entry{
		Timestamp:  "2026-06-03T10:30:45.123456+03:00",
		Request:    `todoz add-task --title "Buy milk"`,
		OK:         true,
		Data:       `{"id":"task-1"}`,
		DurationMS: 1,
	})
	if err != nil {
		t.Fatalf("Record: %v", err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	out := string(b)
	for _, want := range []string{
		"[2026-06-03T10:30:45.123456+03:00]",
		`REQUEST: todoz add-task --title "Buy milk"`,
		"RESPONSE: ok=true",
		`DATA: {"id":"task-1"}`,
		"DURATION: 1ms",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("log missing %q in:\n%s", want, out)
		}
	}
}

func TestRecordFailure(t *testing.T) {
	path := filepath.Join(t.TempDir(), "logs", "app.log")
	_ = Record(path, Entry{
		Timestamp:  "T",
		Request:    "todoz delete-task x",
		OK:         false,
		Error:      "task_not_found",
		Message:    "Task x does not exist",
		DurationMS: 0,
	})
	b, _ := os.ReadFile(path)
	out := string(b)
	if !strings.Contains(out, "ERROR: task_not_found") || !strings.Contains(out, "MESSAGE: Task x does not exist") {
		t.Fatalf("failure block wrong:\n%s", out)
	}
}
