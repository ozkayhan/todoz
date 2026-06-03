package store

import (
	"path/filepath"
	"testing"
)

func TestLockExclusive(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".lock")
	l1, err := Acquire(path, 50)
	if err != nil {
		t.Fatalf("first acquire failed: %v", err)
	}
	if _, err := Acquire(path, 50); err == nil {
		t.Fatal("second acquire should fail while lock held")
	}
	if err := l1.Release(); err != nil {
		t.Fatalf("release failed: %v", err)
	}
	l2, err := Acquire(path, 50)
	if err != nil {
		t.Fatalf("acquire after release failed: %v", err)
	}
	_ = l2.Release()
}
