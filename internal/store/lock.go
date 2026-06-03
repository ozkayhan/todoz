package store

import (
	"os"
	"path/filepath"
	"time"
)

// Lock represents an acquired exclusive lock on a file.
type Lock struct {
	path string
}

// Acquire acquires an exclusive lock on the file at path, retrying for up to
// timeoutMS milliseconds. Returns an error if the lock cannot be acquired
// within the timeout.
func Acquire(path string, timeoutMS int) (*Lock, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	deadline := time.Now().Add(time.Duration(timeoutMS) * time.Millisecond)
	for {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err == nil {
			_ = f.Close()
			return &Lock{path: path}, nil
		}
		if !os.IsExist(err) {
			return nil, err
		}
		if time.Now().After(deadline) {
			return nil, err
		}
		time.Sleep(2 * time.Millisecond)
	}
}

// Release releases the lock by removing the lock file.
func (l *Lock) Release() error {
	return os.Remove(l.path)
}
