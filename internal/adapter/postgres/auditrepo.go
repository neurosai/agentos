package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/neurosai/agentos/internal/domain/audit"
	"github.com/neurosai/agentos/internal/port"
)

// AuditRepository persists audit events.
type AuditRepository struct {
	pool *Pool
}

func NewAuditRepository(pool *Pool) *AuditRepository {
	return &AuditRepository{pool: pool}
}

func (r *AuditRepository) Insert(ctx context.Context, event audit.Event) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO audit_events (
			event_id, occurred_at, tenant_id, subject_id, agent_id, task_id,
			event_type, resource_type, resource_id, action, decision, status,
			payload_hash, prev_hash, event_hash, trace_id, span_id
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`,
		event.EventID, event.OccurredAt, event.TenantID, event.SubjectID, event.AgentID, event.TaskID,
		event.EventType, event.ResourceType, event.ResourceID, event.Action, event.Decision, event.Status,
		event.PayloadHash, event.PrevHash, event.EventHash, event.TraceID, event.SpanID)
	return err
}

func (r *AuditRepository) GetByID(ctx context.Context, eventID string) (audit.Event, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT event_id, occurred_at, tenant_id, subject_id, agent_id, task_id,
			event_type, resource_type, resource_id, action, decision, status,
			payload_hash, prev_hash, event_hash, trace_id, span_id
		FROM audit_events WHERE event_id = $1`, eventID)
	return scanEvent(row)
}

func (r *AuditRepository) List(ctx context.Context, filter port.AuditQueryFilter) ([]audit.Event, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 100
	}
	var rows pgx.Rows
	var err error
	switch {
	case filter.TaskID != "":
		rows, err = r.pool.Query(ctx, `
			SELECT event_id, occurred_at, tenant_id, subject_id, agent_id, task_id,
				event_type, resource_type, resource_id, action, decision, status,
				payload_hash, prev_hash, event_hash, trace_id, span_id
			FROM audit_events WHERE task_id = $1 ORDER BY occurred_at LIMIT $2`, filter.TaskID, limit)
	case filter.TenantID != "":
		rows, err = r.pool.Query(ctx, `
			SELECT event_id, occurred_at, tenant_id, subject_id, agent_id, task_id,
				event_type, resource_type, resource_id, action, decision, status,
				payload_hash, prev_hash, event_hash, trace_id, span_id
			FROM audit_events WHERE tenant_id = $1 ORDER BY occurred_at DESC LIMIT $2`, filter.TenantID, limit)
	default:
		rows, err = r.pool.Query(ctx, `
			SELECT event_id, occurred_at, tenant_id, subject_id, agent_id, task_id,
				event_type, resource_type, resource_id, action, decision, status,
				payload_hash, prev_hash, event_hash, trace_id, span_id
			FROM audit_events ORDER BY occurred_at DESC LIMIT $1`, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEvents(rows)
}

func (r *AuditRepository) LastHash(ctx context.Context, streamID string) (string, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT event_hash FROM audit_events WHERE tenant_id = $1
		ORDER BY occurred_at DESC LIMIT 1`, streamID)
	var hash string
	err := row.Scan(&hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	return hash, err
}

func (r *AuditRepository) ListByTrace(ctx context.Context, traceID string) ([]audit.Event, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT event_id, occurred_at, tenant_id, subject_id, agent_id, task_id,
			event_type, resource_type, resource_id, action, decision, status,
			payload_hash, prev_hash, event_hash, trace_id, span_id
		FROM audit_events WHERE trace_id = $1 ORDER BY occurred_at`, traceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEvents(rows)
}

var _ port.AuditRepository = (*AuditRepository)(nil)

func scanEvent(row pgx.Row) (audit.Event, error) {
	var e audit.Event
	err := row.Scan(
		&e.EventID, &e.OccurredAt, &e.TenantID, &e.SubjectID, &e.AgentID, &e.TaskID,
		&e.EventType, &e.ResourceType, &e.ResourceID, &e.Action, &e.Decision, &e.Status,
		&e.PayloadHash, &e.PrevHash, &e.EventHash, &e.TraceID, &e.SpanID,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return audit.Event{}, fmt.Errorf("audit event not found")
	}
	return e, err
}

func scanEvents(rows pgx.Rows) ([]audit.Event, error) {
	var out []audit.Event
	for rows.Next() {
		var e audit.Event
		if err := rows.Scan(
			&e.EventID, &e.OccurredAt, &e.TenantID, &e.SubjectID, &e.AgentID, &e.TaskID,
			&e.EventType, &e.ResourceType, &e.ResourceID, &e.Action, &e.Decision, &e.Status,
			&e.PayloadHash, &e.PrevHash, &e.EventHash, &e.TraceID, &e.SpanID,
		); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}
