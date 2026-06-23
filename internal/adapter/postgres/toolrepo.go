package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/neurosai/agentos/internal/domain/tool"
	"github.com/neurosai/agentos/internal/port"
)

// ToolInvocationRepository persists tool invocations.
type ToolInvocationRepository struct {
	pool *Pool
}

func NewToolInvocationRepository(pool *Pool) *ToolInvocationRepository {
	return &ToolInvocationRepository{pool: pool}
}

func (r *ToolInvocationRepository) Create(ctx context.Context, rec tool.InvocationRecord) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO tool_invocations (id, tool_id, task_id, agent_id, status, audit_event_id, started_at, completed_at, idempotency_key)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		rec.ID, rec.ToolID, rec.TaskID, rec.AgentID, rec.Status, rec.AuditEventID,
		rec.StartedAt, rec.CompletedAt, nullIfEmpty(rec.IdempotencyKey))
	return err
}

func (r *ToolInvocationRepository) Get(ctx context.Context, callID string) (tool.InvocationRecord, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, tool_id, task_id, agent_id, status, audit_event_id, started_at, completed_at, idempotency_key
		FROM tool_invocations WHERE id = $1`, callID)
	return scanInvocation(row)
}

func (r *ToolInvocationRepository) GetByIdempotency(ctx context.Context, toolID, key string) (tool.InvocationRecord, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, tool_id, task_id, agent_id, status, audit_event_id, started_at, completed_at, idempotency_key
		FROM tool_invocations WHERE tool_id = $1 AND idempotency_key = $2`, toolID, key)
	return scanInvocation(row)
}

var _ port.ToolInvocationRepository = (*ToolInvocationRepository)(nil)

func scanInvocation(row pgx.Row) (tool.InvocationRecord, error) {
	var rec tool.InvocationRecord
	var idem *string
	err := row.Scan(&rec.ID, &rec.ToolID, &rec.TaskID, &rec.AgentID, &rec.Status,
		&rec.AuditEventID, &rec.StartedAt, &rec.CompletedAt, &idem)
	if errors.Is(err, pgx.ErrNoRows) {
		return tool.InvocationRecord{}, fmt.Errorf("invocation not found")
	}
	if idem != nil {
		rec.IdempotencyKey = *idem
	}
	return rec, err
}

func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
