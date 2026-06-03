package model

import "testing"

func TestComputeOverdue(t *testing.T) {
	today := "2026-06-03"
	cases := []struct {
		name   string
		task   Task
		expect bool
	}{
		{"pending past date is overdue", Task{Status: StatusPending, Date: "2026-06-01"}, true},
		{"pending today is not overdue", Task{Status: StatusPending, Date: "2026-06-03"}, false},
		{"pending future is not overdue", Task{Status: StatusPending, Date: "2026-06-10"}, false},
		{"completed past is not overdue", Task{Status: StatusCompleted, Date: "2026-06-01"}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := ComputeOverdue(c.task, today); got != c.expect {
				t.Fatalf("ComputeOverdue=%v, want %v", got, c.expect)
			}
		})
	}
}
