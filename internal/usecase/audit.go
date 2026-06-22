package usecase

import (
	"context"

	"github.com/neurosai/agentos/internal/domain/audit"
)

// AuditService manages append-only audit trails.
//
// Pre: internal callers only for Append; audit.read for Query/Proof.
// Post: hash chain extended; optional anchor recorded.
type AuditService interface {
	Record(ctx context.Context, event audit.Event) error
	Query(ctx context.Context, tenantID, taskID string, limit int) ([]audit.Event, error)
	Get(ctx context.Context, eventID string) (audit.Event, error)
	Proof(ctx context.Context, streamID string) (audit.Proof, error)
	Anchor(ctx context.Context, streamID string) (audit.Anchor, error)
}
