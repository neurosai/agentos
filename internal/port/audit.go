package port

import (
	"context"

	"github.com/neurosai/agentos/internal/domain/audit"
)

// AuditSink appends and queries tamper-evident audit events.
type AuditSink interface {
	Append(ctx context.Context, event audit.Event) error
	Get(ctx context.Context, eventID string) (audit.Event, error)
	Query(ctx context.Context, filter AuditQueryFilter) ([]audit.Event, error)
	Proof(ctx context.Context, streamID string) (audit.Proof, error)
	Anchor(ctx context.Context, anchor audit.Anchor) error
}

// AuditQueryFilter filters audit event searches.
type AuditQueryFilter struct {
	TenantID  string
	TaskID    string
	SubjectID string
	AgentID   string
	Limit     int
	Cursor    string
}

// AuditRepository persists audit events.
type AuditRepository interface {
	Insert(ctx context.Context, event audit.Event) error
	GetByID(ctx context.Context, eventID string) (audit.Event, error)
	List(ctx context.Context, filter AuditQueryFilter) ([]audit.Event, error)
	LastHash(ctx context.Context, streamID string) (string, error)
}
