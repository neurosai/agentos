package usecase

import (
	"context"

	"github.com/neurosai/agentos/internal/domain/task"
)

// CreateTaskInput is the northbound task submission request.
type CreateTaskInput struct {
	TenantID  string
	AgentRef  string
	ContextID string
	Input     map[string]any
	Labels    map[string]string
	CreatedBy string
}

// TaskService coordinates task lifecycle.
//
// Pre: policy allows task.submit for subject.
// Post: task persisted in CREATED/ACCEPTED, audit event task.create appended.
type TaskService interface {
	Create(ctx context.Context, in CreateTaskInput) (task.Task, error)
	Get(ctx context.Context, taskID string) (task.Task, error)
	Cancel(ctx context.Context, taskID string) (task.Task, error)
	AddMessage(ctx context.Context, taskID string, msg task.Message) error
	RequestApproval(ctx context.Context, taskID string, approval task.Approval) error
	DecideApproval(ctx context.Context, taskID string, approvalID string, approved bool, decidedBy string) (task.Task, error)
	ListArtifacts(ctx context.Context, taskID string) ([]task.Artifact, error)
}

// TaskEventStreamer exposes task event streams for SSE/gRPC adapters.
type TaskEventStreamer interface {
	StreamEvents(ctx context.Context, taskID string, fromBeginning bool) (<-chan task.Message, error)
}
