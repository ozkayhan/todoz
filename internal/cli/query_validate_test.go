package cli

import "testing"

func TestValidateQueryFlags_DaysBackAndAfterDate_Conflict(t *testing.T) {
	flags := map[string]string{"days-back": "7", "after-date": "2026-05-01"}
	if err := ValidateQueryFlags(flags); err == nil {
		t.Fatal("expected conflict error")
	}
}

func TestValidateQueryFlags_NoTrashAndTrashOnly_Conflict(t *testing.T) {
	flags := map[string]string{"no-trash": "true", "trash-only": "true"}
	if err := ValidateQueryFlags(flags); err == nil {
		t.Fatal("expected conflict error")
	}
}

func TestValidateQueryFlags_ValidCombination(t *testing.T) {
	flags := map[string]string{"days-back": "7", "status": "pending", "sort-by": "date", "group-by": "list"}
	if err := ValidateQueryFlags(flags); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateQueryFlags_InvalidStatus(t *testing.T) {
	flags := map[string]string{"status": "overdue"}
	if err := ValidateQueryFlags(flags); err == nil {
		t.Fatal("expected error for invalid status")
	}
}

func TestValidateQueryFlags_InvalidSortBy(t *testing.T) {
	flags := map[string]string{"sort-by": "priority"}
	if err := ValidateQueryFlags(flags); err == nil {
		t.Fatal("expected error for invalid sort-by value")
	}
}
