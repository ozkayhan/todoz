package main

import (
	"os/exec"
	"strings"
	"testing"
)

func TestBinaryAddListEndToEnd(t *testing.T) {
	dir := t.TempDir()
	bin := dir + "/todoz"
	if out, err := exec.Command("go", "build", "-o", bin, ".").CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
	cmd := exec.Command(bin, "add-list", "--name", "Groceries")
	cmd.Env = append(cmd.Environ(), "TODO_LIB_HOME="+dir+"/data", "TODO_APP_NAME=e2e")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), `"ok":true`) {
		t.Fatalf("unexpected output: %s", out)
	}
}
