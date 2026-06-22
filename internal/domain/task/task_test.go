package task

import "testing"

func TestCanTransition(t *testing.T) {
	t.Parallel()

	cases := []struct {
		from, to string
		want     bool
	}{
		{string(StatusCreated), string(StatusAccepted), true},
		{string(StatusCreated), string(StatusRunning), false},
		{string(StatusRunning), string(StatusWaitingApproval), true},
		{string(StatusRunning), string(StatusCompleted), true},
		{string(StatusCompleted), string(StatusRunning), false},
		{string(StatusWaitingInput), string(StatusRunning), true},
		{string(StatusFailed), string(StatusCancelled), false},
	}
	for _, tc := range cases {
		got := CanTransition(Status(tc.from), Status(tc.to))
		if got != tc.want {
			t.Fatalf("%s -> %s: got %v want %v", tc.from, tc.to, got, tc.want)
		}
	}
}

func TestTransition(t *testing.T) {
	t.Parallel()

	task := &Task{ID: "task_01", Status: StatusCreated}
	if err := Transition(task, StatusRunning); err == nil {
		t.Fatal("expected error for CREATED -> RUNNING")
	}
	if err := Transition(task, StatusAccepted); err != nil {
		t.Fatalf("transition: %v", err)
	}
	if task.Status != StatusAccepted {
		t.Fatalf("status %s", task.Status)
	}
}

func TestTerminal(t *testing.T) {
	t.Parallel()

	if !StatusCompleted.Terminal() {
		t.Fatal("completed should be terminal")
	}
	if StatusRunning.Terminal() {
		t.Fatal("running should not be terminal")
	}
}
