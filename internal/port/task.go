package port

import (
	"context"

	"github.com/neurosai/agentos/internal/domain/task"
)

// TaskRepository persists task aggregates.
type TaskRepository interface {
	Create(ctx context.Context, t task.Task) error
	Get(ctx context.Context, id string) (task.Task, error)
	UpdateStatus(ctx context.Context, id string, status task.Status) error
	ListByTenant(ctx context.Context, tenantID string, limit int) ([]task.Task, error)
}

// TaskMessageRepository stores task messages.
type TaskMessageRepository interface {
	Append(ctx context.Context, msg task.Message) error
	List(ctx context.Context, taskID string) ([]task.Message, error)
}

// TaskArtifactRepository stores task artifacts.
type TaskArtifactRepository interface {
	Create(ctx context.Context, artifact task.Artifact) error
	List(ctx context.Context, taskID string) ([]task.Artifact, error)
}

// TaskApprovalRepository stores approval records.
type TaskApprovalRepository interface {
	Create(ctx context.Context, approval task.Approval) error
	GetPending(ctx context.Context, taskID string) (task.Approval, error)
	Decide(ctx context.Context, approvalID string, approved bool, decidedBy string) error
}

// AgentRunRepository tracks agent execution attempts.
type AgentRunRepository interface {
	Create(ctx context.Context, run task.AgentRun) error
	Update(ctx context.Context, run task.AgentRun) error
	GetByTask(ctx context.Context, taskID string) ([]task.AgentRun, error)
}
