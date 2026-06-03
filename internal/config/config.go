// Package config resolves runtime configuration from environment variables
// and derives every filesystem path todoz uses.
//
// Environment variables:
//
//	TODO_LIB_HOME  Root directory for all todoz data. Default: <user config>/todoz.
//	TODO_APP_NAME  Identifier of the calling app, used in the operation log.
//	               Default: "unknown".
//
// Keeping all path construction in one place means the rest of the codebase
// never hard-codes a filename.
package config

import (
	"os"
	"path/filepath"
)

// Config is the resolved configuration for a single todoz invocation.
type Config struct {
	// Home is the root directory holding the event log, backups, and logs.
	Home string
	// AppName identifies the calling application in the operation log.
	AppName string
}

// Load reads configuration from the environment, applying defaults.
func Load() (Config, error) {
	home := os.Getenv("TODO_LIB_HOME")
	if home == "" {
		base, err := os.UserConfigDir()
		if err != nil {
			return Config{}, err
		}
		home = filepath.Join(base, "todoz")
	}

	app := os.Getenv("TODO_APP_NAME")
	if app == "" {
		app = "unknown"
	}

	return Config{Home: home, AppName: app}, nil
}

// EventsPath is the append-only event log file.
func (c Config) EventsPath() string { return filepath.Join(c.Home, "events.jsonl") }

// BackupPath is the pre-compaction backup of the event log.
func (c Config) BackupPath() string { return filepath.Join(c.Home, "events.jsonl.backup") }

// LockPath is the exclusive lock file guarding event-log appends.
func (c Config) LockPath() string { return filepath.Join(c.Home, ".lock") }

// LogsDir is the directory holding per-app operation logs.
func (c Config) LogsDir() string { return filepath.Join(c.Home, "logs") }

// AppLogPath is the operation log for the current app.
func (c Config) AppLogPath() string { return filepath.Join(c.LogsDir(), c.AppName+".log") }
