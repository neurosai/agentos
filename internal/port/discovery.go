package port

import (
	"context"

	"github.com/neurosai/agentos/internal/domain/discovery"
)

// DiscoveryRepository persists discovery jobs and observations.
type DiscoveryRepository interface {
	CreateJob(ctx context.Context, job discovery.Job) error
	UpdateJob(ctx context.Context, job discovery.Job) error
	GetJob(ctx context.Context, id string) (discovery.Job, error)
	AppendObservations(ctx context.Context, observations []discovery.Observation) error
	ListObservations(ctx context.Context, jobID string) ([]discovery.Observation, error)
}

// Collector runs a safe read-only discovery collector.
type Collector interface {
	Kind() discovery.CollectorKind
	Collect(ctx context.Context, job discovery.Job) ([]discovery.Observation, error)
}
