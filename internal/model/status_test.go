package model

import "testing"

// Only pending and completed are valid; overdue is NOT a status.
func TestStatusValid(t *testing.T) {
	if !StatusPending.Valid() {
		t.Fatal("pending should be valid")
	}
	if !StatusCompleted.Valid() {
		t.Fatal("completed should be valid")
	}
	if Status("overdue").Valid() {
		t.Fatal("overdue must NOT be a valid status")
	}
	if Status("").Valid() {
		t.Fatal("empty status must be invalid")
	}
}
