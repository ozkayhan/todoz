package cli

import (
	"strings"
	"testing"
)

func TestRunDispatchesAddList(t *testing.T) {
	t.Setenv("TODO_LIB_HOME", t.TempDir())
	t.Setenv("TODO_APP_NAME", "test")
	out, code := Run([]string{"add-list", "--name", "Groceries"})
	if code != 0 {
		t.Fatalf("exit code = %d, want 0; out=%s", code, out)
	}
	if !strings.Contains(out, `"ok":true`) {
		t.Fatalf("unexpected output: %s", out)
	}
}

func TestRunUnknownCommand(t *testing.T) {
	t.Setenv("TODO_LIB_HOME", t.TempDir())
	out, code := Run([]string{"frobnicate"})
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if !strings.Contains(out, "unknown_command") {
		t.Fatalf("want unknown_command, got %s", out)
	}
}

func TestRunNoCommand(t *testing.T) {
	t.Setenv("TODO_LIB_HOME", t.TempDir())
	_, code := Run([]string{})
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
}
