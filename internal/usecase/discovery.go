package usecase

import (
	"context"

	"github.com/neurosai/agentos/internal/domain/discovery"
)

// DiscoveryService runs safe collectors and routes observations to CatalogD.
//
// Pre: collector must be in safe allowlist; mode read_only; policy discovery.request.
// Post: observations written to catalog; audit discovery.collect appended.
// Constraint: no direct write to MemoryD without CatalogD (enforced in domain).
type DiscoveryService interface {
	CreateJob(ctx context.Context, job discovery.Job) (discovery.Job, error)
	GetJob(ctx context.Context, jobID string) (discovery.Job, error)
	ApproveJob(ctx context.Context, jobID string, approvedBy string) (discovery.Job, error)
	ListCollectors(ctx context.Context) ([]discovery.CollectorKind, error)
	ListObservations(ctx context.Context, jobID string) ([]discovery.Observation, error)
}
