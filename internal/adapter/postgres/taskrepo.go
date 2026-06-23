package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/neurosai/agentos/internal/domain/task"
	"github.com/neurosai/agentos/internal/port"
)

// TaskRepository persists tasks in PostgreSQL.
type TaskRepository struct {
	pool *Pool
}

func NewTaskRepository(pool *Pool) *TaskRepository {
	return &TaskRepository{pool: pool}
}

func (r *TaskRepository) Create(ctx context.Context, t task.Task) error {
	input, _ := json.Marshal(t.Input)
	labels, _ := json.Marshal(t.Labels)
	_, err := r.pool.Exec(ctx, `
		INSERT INTO tasks (id, tenant_id, context_id, agent_ref, status, input, labels, created_by, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		t.ID, t.TenantID, t.ContextID, t.AgentRef, string(t.Status), input, labels, t.CreatedBy, t.CreatedAt, t.UpdatedAt)
	return err
}

func (r *TaskRepository) Get(ctx context.Context, id string) (task.Task, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, context_id, agent_ref, status, input, labels, created_by, created_at, updated_at
		FROM tasks WHERE id = $1`, id)
	var t task.Task
	var status string
	var input, labels []byte
	err := row.Scan(&t.ID, &t.TenantID, &t.ContextID, &t.AgentRef, &status, &input, &labels, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return task.Task{}, fmt.Errorf("task not found: %s", id)
	}
	if err != nil {
		return task.Task{}, err
	}
	t.Status = task.Status(status)
	_ = json.Unmarshal(input, &t.Input)
	_ = json.Unmarshal(labels, &t.Labels)
	return t, nil
}

func (r *TaskRepository) UpdateStatus(ctx context.Context, id string, status task.Status) error {
	_, err := r.pool.Exec(ctx, `UPDATE tasks SET status = $1, updated_at = $2 WHERE id = $3`,
		string(status), time.Now().UTC(), id)
	return err
}

func (r *TaskRepository) ListByTenant(ctx context.Context, tenantID string, limit int) ([]task.Task, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, context_id, agent_ref, status, input, labels, created_by, created_at, updated_at
		FROM tasks WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT $2`, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []task.Task
	for rows.Next() {
		var t task.Task
		var status string
		var input, labels []byte
		if err := rows.Scan(&t.ID, &t.TenantID, &t.ContextID, &t.AgentRef, &status, &input, &labels, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		t.Status = task.Status(status)
		_ = json.Unmarshal(input, &t.Input)
		_ = json.Unmarshal(labels, &t.Labels)
		out = append(out, t)
	}
	return out, rows.Err()
}

var _ port.TaskRepository = (*TaskRepository)(nil)

// TaskMessageRepository stores messages in PostgreSQL.
type TaskMessageRepository struct {
	pool *Pool
}

func NewTaskMessageRepository(pool *Pool) *TaskMessageRepository {
	return &TaskMessageRepository{pool: pool}
}

func (r *TaskMessageRepository) Append(ctx context.Context, msg task.Message) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO task_messages (id, task_id, role, content, created_at)
		VALUES ($1,$2,$3,$4,$5)`,
		msg.ID, msg.TaskID, msg.Role, msg.Content, msg.CreatedAt)
	return err
}

func (r *TaskMessageRepository) List(ctx context.Context, taskID string) ([]task.Message, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, task_id, role, content, created_at FROM task_messages
		WHERE task_id = $1 ORDER BY created_at`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []task.Message
	for rows.Next() {
		var m task.Message
		if err := rows.Scan(&m.ID, &m.TaskID, &m.Role, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

var _ port.TaskMessageRepository = (*TaskMessageRepository)(nil)

// TaskArtifactRepository stores artifacts.
type TaskArtifactRepository struct {
	pool *Pool
}

func NewTaskArtifactRepository(pool *Pool) *TaskArtifactRepository {
	return &TaskArtifactRepository{pool: pool}
}

func (r *TaskArtifactRepository) Create(ctx context.Context, a task.Artifact) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO task_artifacts (id, task_id, name, content_type, uri, size_bytes, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		a.ID, a.TaskID, a.Name, a.ContentType, a.URI, a.SizeBytes, a.CreatedAt)
	return err
}

func (r *TaskArtifactRepository) List(ctx context.Context, taskID string) ([]task.Artifact, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, task_id, name, content_type, uri, size_bytes, created_at
		FROM task_artifacts WHERE task_id = $1`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []task.Artifact
	for rows.Next() {
		var a task.Artifact
		if err := rows.Scan(&a.ID, &a.TaskID, &a.Name, &a.ContentType, &a.URI, &a.SizeBytes, &a.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

var _ port.TaskArtifactRepository = (*TaskArtifactRepository)(nil)

// TaskApprovalRepository stores approvals.
type TaskApprovalRepository struct {
	pool *Pool
}

func NewTaskApprovalRepository(pool *Pool) *TaskApprovalRepository {
	return &TaskApprovalRepository{pool: pool}
}

func (r *TaskApprovalRepository) Create(ctx context.Context, a task.Approval) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO task_approvals (id, task_id, requested, decided_at, decided_by, approved, reason)
		VALUES ($1,$2,$3,NULL,NULL,NULL,$4)`,
		a.ID, a.TaskID, a.Requested, a.Reason)
	return err
}

func (r *TaskApprovalRepository) GetPending(ctx context.Context, taskID string) (task.Approval, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, task_id, requested, decided_at, decided_by, approved, reason
		FROM task_approvals WHERE task_id = $1 AND decided_at IS NULL
		ORDER BY requested DESC LIMIT 1`, taskID)
	var a task.Approval
	var decidedAt *time.Time
	var decidedBy *string
	var approved *bool
	err := row.Scan(&a.ID, &a.TaskID, &a.Requested, &decidedAt, &decidedBy, &approved, &a.Reason)
	if errors.Is(err, pgx.ErrNoRows) {
		return task.Approval{}, fmt.Errorf("pending approval not found")
	}
	if err != nil {
		return task.Approval{}, err
	}
	if decidedBy != nil {
		a.DecidedBy = *decidedBy
	}
	if approved != nil {
		a.Approved = *approved
	}
	a.DecidedAt = decidedAt
	return a, nil
}

func (r *TaskApprovalRepository) Decide(ctx context.Context, approvalID string, approved bool, decidedBy string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE task_approvals SET decided_at = $1, decided_by = $2, approved = $3
		WHERE id = $4`, time.Now().UTC(), decidedBy, approved, approvalID)
	return err
}

var _ port.TaskApprovalRepository = (*TaskApprovalRepository)(nil)

// AgentRunRepository tracks agent runs.
type AgentRunRepository struct {
	pool *Pool
}

func NewAgentRunRepository(pool *Pool) *AgentRunRepository {
	return &AgentRunRepository{pool: pool}
}

func (r *AgentRunRepository) Create(ctx context.Context, run task.AgentRun) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO agent_runs (id, task_id, agent_ref, status, started_at, ended_at)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		run.ID, run.TaskID, run.AgentRef, string(run.Status), run.StartedAt, run.EndedAt)
	return err
}

func (r *AgentRunRepository) Update(ctx context.Context, run task.AgentRun) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE agent_runs SET status = $1, ended_at = $2 WHERE id = $3`,
		string(run.Status), run.EndedAt, run.ID)
	return err
}

func (r *AgentRunRepository) GetByTask(ctx context.Context, taskID string) ([]task.AgentRun, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, task_id, agent_ref, status, started_at, ended_at FROM agent_runs WHERE task_id = $1`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []task.AgentRun
	for rows.Next() {
		var run task.AgentRun
		var status string
		if err := rows.Scan(&run.ID, &run.TaskID, &run.AgentRef, &status, &run.StartedAt, &run.EndedAt); err != nil {
			return nil, err
		}
		run.Status = task.Status(status)
		out = append(out, run)
	}
	return out, rows.Err()
}

var _ port.AgentRunRepository = (*AgentRunRepository)(nil)
