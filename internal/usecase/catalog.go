package usecase

import (
	"context"

	"github.com/neurosai/agentos/internal/domain/catalog"
	"github.com/neurosai/agentos/internal/domain/discovery"
)

// CatalogService manages the typed operational graph.
//
// Pre: declared entities may be upserted by catalog.write principals.
// Post: observed edges materialized separately from declared metadata.
type CatalogService interface {
	UpsertEntity(ctx context.Context, entity catalog.Entity) (catalog.Entity, error)
	GetEntity(ctx context.Context, ref string) (catalog.Entity, error)
	Graph(ctx context.Context, rootRef string, depth int) (catalog.GraphSlice, error)
	MergeObservations(ctx context.Context, observations []discovery.Observation) (accepted, rejected int, err error)
	Search(ctx context.Context, query string, limit int) ([]catalog.Entity, error)
}
