package task

import (
	"fmt"
	"time"
)

// Status represents the task lifecycle state.
type Status string

const (
	StatusCreated          Status = "CREATED"
	StatusAccepted         Status = "ACCEPTED"
	StatusRunning          Status = "RUNNING"
	StatusWaitingInput     Status = "WAITING_INPUT"
	StatusWaitingApproval  Status = "WAITING_APPROVAL"
	StatusCompleted        Status = "COMPLETED"
	StatusFailed           Status = "FAILED"
	StatusCancelled        Status = "CANCELLED"
)

// Terminal returns true when no further transitions are allowed.
func (s Status) Terminal() bool {
	return s == StatusCompleted || s == StatusFailed || s == StatusCancelled
}

// Task is the aggregate root for task execution.
type Task struct {
	ID         string
	TenantID   string
	ContextID  string
	AgentRef   string
	Status     Status
	Input      map[string]any
	Labels     map[string]string
	CreatedBy  string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Message is a conversational turn within a task.
type Message struct {
	ID        string
	TaskID    string
	Role      string
	Content   string
	CreatedAt time.Time
}

// Artifact is an output produced by task execution.
type Artifact struct {
	ID          string
	TaskID      string
	Name        string
	ContentType string
	URI         string
	SizeBytes   int64
	CreatedAt   time.Time
}

// Approval records a human approval decision.
type Approval struct {
	ID         string
	TaskID     string
	Requested  time.Time
	DecidedAt  *time.Time
	DecidedBy  string
	Approved   bool
	Reason     string
}

// AgentRun links a task to a concrete agent execution attempt.
type AgentRun struct {
	ID        string
	TaskID    string
	AgentRef  string
	Status    Status
	StartedAt time.Time
	EndedAt   *time.Time
}

// CanTransition reports whether moving from current to next is valid.
func CanTransition(from, to Status) bool {
	if from == to {
		return true
	}
	if from.Terminal() {
		return false
	}
	allowed, ok := transitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

// Transition applies a status change or returns an error.
func Transition(t *Task, to Status) error {
	if t == nil {
		return fmt.Errorf("nil task")
	}
	if !CanTransition(t.Status, to) {
		return fmt.Errorf("invalid transition %s -> %s", t.Status, to)
	}
	t.Status = to
	return nil
}

var transitions = map[Status][]Status{
	StatusCreated:         {StatusAccepted, StatusCancelled},
	StatusAccepted:        {StatusRunning, StatusCancelled},
	StatusRunning:         {StatusWaitingInput, StatusWaitingApproval, StatusCompleted, StatusFailed, StatusCancelled},
	StatusWaitingInput:    {StatusRunning, StatusCancelled, StatusFailed},
	StatusWaitingApproval: {StatusRunning, StatusCancelled, StatusFailed},
	StatusCompleted:       {},
	StatusFailed:          {},
	StatusCancelled:       {},
}
