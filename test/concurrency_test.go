package test

import (
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"testing"
)

func TestConcurrentAppendsNoLoss(t *testing.T) {
	dir := t.TempDir()
	bin := dir + "/todoz"
	if out, err := exec.Command("go", "build", "-o", bin, "../cmd/todoz").CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	home := dir + "/data"
	env := func() []string {
		return append([]string{}, "TODO_LIB_HOME="+home, "TODO_APP_NAME=conc")
	}

	seed := exec.Command(bin, "add-list", "--name", "L")
	seed.Env = env()
	seedOut, err := seed.CombinedOutput()
	if err != nil {
		t.Fatalf("seed: %v\n%s", err, seedOut)
	}
	listID := extractID(t, string(seedOut))

	const N = 50
	var wg sync.WaitGroup
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			c := exec.Command(bin, "add-task", "--title", "t"+strconv.Itoa(i), "--date", "2026-06-05", "--list", listID)
			c.Env = env()
			if out, err := c.CombinedOutput(); err != nil {
				t.Errorf("worker %d failed: %v\n%s", i, err, out)
			}
		}(i)
	}
	wg.Wait()

	load := exec.Command(bin, "load", "--list", listID)
	load.Env = env()
	out, err := load.CombinedOutput()
	if err != nil {
		t.Fatalf("load: %v\n%s", err, out)
	}
	got := strings.Count(string(out), `"id":"task-`)
	if got != N {
		t.Fatalf("expected %d tasks after concurrent appends, got %d\n%s", N, got, out)
	}
}

func extractID(t *testing.T, s string) string {
	t.Helper()
	marker := `"id":"`
	i := strings.Index(s, marker)
	if i < 0 {
		t.Fatalf("no id in %s", s)
	}
	rest := s[i+len(marker):]
	j := strings.Index(rest, `"`)
	return rest[:j]
}
