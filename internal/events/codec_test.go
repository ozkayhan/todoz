package events

import "testing"

func TestEncodeDecodeRoundTrip(t *testing.T) {
	in := Event{
		Type:   TypeTaskCreated,
		At:     "2026-06-03T10:00:00.000000+03:00",
		TaskID: "task-1",
		ListID: "list-1",
		Title:  "Buy milk",
		Date:   "2026-06-05",
	}
	line, err := Encode(in)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if len(line) > 0 && line[len(line)-1] == '\n' {
		t.Fatal("Encode must NOT include trailing newline")
	}
	out, err := Decode(line)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if out.Type != in.Type || out.At != in.At || out.TaskID != in.TaskID ||
		out.ListID != in.ListID || out.Title != in.Title || out.Date != in.Date {
		t.Fatalf("round trip mismatch:\n got %+v\nwant %+v", out, in)
	}
}

func TestDecodeRejectsGarbage(t *testing.T) {
	if _, err := Decode("{not json"); err == nil {
		t.Fatal("expected error decoding garbage")
	}
}
