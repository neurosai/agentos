package memory

import (
	"context"
	"sync"
	"time"

	"github.com/neurosai/agentos/internal/domain/task"
	"github.com/neurosai/agentos/internal/port"
)

// TaskRepository is an in-memory TaskRepository.
type TaskRepository struct {
	mu    sync.RWMutex
	tasks map[string]task.Task
}

func NewTaskRepository() *TaskRepository {
	return &TaskRepository{tasks: make(map[string]task.Task)}
}

func (r *TaskRepository) Create(ctx context.Context, t task.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tasks[t.ID] = t
	return nil
}

func (r *TaskRepository) Get(ctx context.Context, id string) (task.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tasks[id]
	if !ok {
		return task.Task{}, errNotFound("task", id)
	}
	return t, nil
}

func (r *TaskRepository) UpdateStatus(ctx context.Context, id string, status task.Status) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.tasks[id]
	if !ok {
		return errNotFound("task", id)
	}
	t.Status = status
	r.tasks[id] = t
	return nil
}

func (r *TaskRepository) ListByTenant(ctx context.Context, tenantID string, limit int) ([]task.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []task.Task
	for _, t := range r.tasks {
		if t.TenantID == tenantID {
			out = append(out, t)
		}
	}
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

var _ port.TaskRepository = (*TaskRepository)(nil)

// TaskMessageRepository stores messages in memory.
type TaskMessageRepository struct {
	mu       sync.RWMutex
	messages map[string][]task.Message
}

func NewTaskMessageRepository() *TaskMessageRepository {
	return &TaskMessageRepository{messages: make(map[string][]task.Message)}
}

func (r *TaskMessageRepository) Append(ctx context.Context, msg task.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.messages[msg.TaskID] = append(r.messages[msg.TaskID], msg)
	return nil
}

func (r *TaskMessageRepository) List(ctx context.Context, taskID string) ([]task.Message, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]task.Message(nil), r.messages[taskID]...), nil
}

var _ port.TaskMessageRepository = (*TaskMessageRepository)(nil)

// TaskArtifactRepository stores artifacts in memory.
type TaskArtifactRepository struct {
	mu        sync.RWMutex
	artifacts map[string][]task.Artifact
}

func NewTaskArtifactRepository() *TaskArtifactRepository {
	return &TaskArtifactRepository{artifacts: make(map[string][]task.Artifact)}
}

func (r *TaskArtifactRepository) Create(ctx context.Context, a task.Artifact) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.artifacts[a.TaskID] = append(r.artifacts[a.TaskID], a)
	return nil
}

func (r *TaskArtifactRepository) List(ctx context.Context, taskID string) ([]task.Artifact, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]task.Artifact(nil), r.artifacts[taskID]...), nil
}

var _ port.TaskArtifactRepository = (*TaskArtifactRepository)(nil)

// TaskApprovalRepository stores approvals in memory.
type TaskApprovalRepository struct {
	mu        sync.RWMutex
	approvals map[string]task.Approval
	byTask    map[string]string
}

func NewTaskApprovalRepository() *TaskApprovalRepository {
	return &TaskApprovalRepository{
		approvals: make(map[string]task.Approval),
		byTask:    make(map[string]string),
	}
}

func (r *TaskApprovalRepository) Create(ctx context.Context, a task.Approval) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.approvals[a.ID] = a
	r.byTask[a.TaskID] = a.ID
	return nil
}

func (r *TaskApprovalRepository) GetPending(ctx context.Context, taskID string) (task.Approval, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.byTask[taskID]
	if !ok {
		return task.Approval{}, errNotFound("approval", taskID)
	}
	a := r.approvals[id]
	if a.DecidedAt != nil {
		return task.Approval{}, errNotFound("pending approval", taskID)
	}
	return a, nil
}

func (r *TaskApprovalRepository) Decide(ctx context.Context, approvalID string, approved bool, decidedBy string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	a, ok := r.approvals[approvalID]
	if !ok {
		return errNotFound("approval", approvalID)
	}
	now := time.Now().UTC()
	a.DecidedAt = &now
	a.Approved = approved
	a.DecidedBy = decidedBy
	r.approvals[approvalID] = a
	delete(r.byTask, a.TaskID)
	return nil
}

var _ port.TaskApprovalRepository = (*TaskApprovalRepository)(nil)

// AgentRunRepository tracks runs in memory.
type AgentRunRepository struct {
	mu   sync.RWMutex
	runs map[string][]task.AgentRun
}

func NewAgentRunRepository() *AgentRunRepository {
	return &AgentRunRepository{runs: make(map[string][]task.AgentRun)}
}

func (r *AgentRunRepository) Create(ctx context.Context, run task.AgentRun) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.runs[run.TaskID] = append(r.runs[run.TaskID], run)
	return nil
}

func (r *AgentRunRepository) Update(ctx context.Context, run task.AgentRun) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	list := r.runs[run.TaskID]
	for i := range list {
		if list[i].ID == run.ID {
			list[i] = run
			r.runs[run.TaskID] = list
			return nil
		}
	}
	return errNotFound("agent run", run.ID)
}

func (r *AgentRunRepository) GetByTask(ctx context.Context, taskID string) ([]task.AgentRun, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]task.AgentRun(nil), r.runs[taskID]...), nil
}

var _ port.AgentRunRepository = (*AgentRunRepository)(nil)

// EventBroadcaster fans out task events for SSE.
type EventBroadcaster struct {
	mu      sync.RWMutex
	chans   map[string][]chan task.Message
	closed  map[string]bool
}

func NewEventBroadcaster() *EventBroadcaster {
	return &EventBroadcaster{
		chans:  make(map[string][]chan task.Message),
		closed: make(map[string]bool),
	}
}

func (b *EventBroadcaster) Subscribe(taskID string) <-chan task.Message {
	ch := make(chan task.Message, 16)
	b.mu.Lock()
	b.chans[taskID] = append(b.chans[taskID], ch)
	b.mu.Unlock()
	return ch
}

func (b *EventBroadcaster) Publish(msg task.Message) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, ch := range b.chans[msg.TaskID] {
		select {
		case ch <- msg:
		default:
		}
	}
}

func errNotFound(kind, id string) error {
	return &notFoundError{kind: kind, id: id}
}

type notFoundError struct {
	kind string
	id   string
}

func (e *notFoundError) Error() string {
	return e.kind + " not found: " + e.id
}
