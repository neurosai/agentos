package port

import (
	"context"

	"github.com/neurosai/agentos/internal/domain/catalog"
)

// CatalogRepository persists catalog entities and edges.
type CatalogRepository interface {
	UpsertEntity(ctx context.Context, entity catalog.Entity) error
	GetEntity(ctx context.Context, ref string) (catalog.Entity, error)
	UpsertRelation(ctx context.Context, rel catalog.Relation) error
	Graph(ctx context.Context, rootRef string, depth int) (catalog.GraphSlice, error)
	Search(ctx context.Context, query string, limit int) ([]catalog.Entity, error)
	SaveSnapshot(ctx context.Context, snap catalog.Snapshot) error
}

// ObservationIngester bulk-ingests discovery observations into the catalog.
type ObservationIngester interface {
	MergeObservations(ctx context.Context, observations []catalog.Entity) (accepted, rejected int, err error)
}
