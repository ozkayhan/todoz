package config

import (
	"path/filepath"
	"testing"
)

// When TODO_LIB_HOME is set, all paths derive from it.
func TestLoadUsesExplicitHome(t *testing.T) {
	t.Setenv("TODO_LIB_HOME", "/tmp/todoz-test")
	t.Setenv("TODO_APP_NAME", "my-app")

	c, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if c.Home != "/tmp/todoz-test" {
		t.Fatalf("Home = %q, want /tmp/todoz-test", c.Home)
	}
	if c.AppName != "my-app" {
		t.Fatalf("AppName = %q, want my-app", c.AppName)
	}
	if c.EventsPath() != filepath.Join("/tmp/todoz-test", "events.jsonl") {
		t.Fatalf("EventsPath = %q", c.EventsPath())
	}
	if c.LockPath() != filepath.Join("/tmp/todoz-test", ".lock") {
		t.Fatalf("LockPath = %q", c.LockPath())
	}
	if c.AppLogPath() != filepath.Join("/tmp/todoz-test", "logs", "my-app.log") {
		t.Fatalf("AppLogPath = %q", c.AppLogPath())
	}
}

// AppName falls back to "unknown" when unset.
func TestLoadDefaultsAppName(t *testing.T) {
	t.Setenv("TODO_LIB_HOME", "/tmp/todoz-test")
	t.Setenv("TODO_APP_NAME", "")

	c, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if c.AppName != "unknown" {
		t.Fatalf("AppName = %q, want unknown", c.AppName)
	}
}
