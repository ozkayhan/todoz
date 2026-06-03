package response

import (
	"encoding/json"
	"testing"
)

func TestSuccessEnvelope(t *testing.T) {
	r := Success(map[string]string{"id": "task-1"})
	if !r.OK {
		t.Fatal("Success.OK must be true")
	}
	b, _ := json.Marshal(r)
	var m map[string]any
	_ = json.Unmarshal(b, &m)
	if m["ok"] != true {
		t.Fatalf("json ok field wrong: %s", b)
	}
	if _, has := m["data"]; !has {
		t.Fatalf("json missing data: %s", b)
	}
}

func TestErrorEnvelope(t *testing.T) {
	r := Error("task_not_found", "Task x does not exist")
	if r.OK {
		t.Fatal("Error.OK must be false")
	}
	if r.ErrCode != "task_not_found" || r.Message != "Task x does not exist" {
		t.Fatalf("error fields wrong: %+v", r)
	}
}
