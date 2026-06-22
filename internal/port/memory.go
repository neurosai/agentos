package port

import (
	"context"

	"github.com/neurosai/agentos/internal/domain/memory"
)

// MemoryRepository persists memory records.
type MemoryRepository interface {
	Create(ctx context.Context, record memory.Record) error
	Get(ctx context.Context, id string) (memory.Record, error)
	SoftDelete(ctx context.Context, id string) error
	Revise(ctx context.Context, id string, revised memory.Record) error
	Query(ctx context.Context, q memory.Query, filters []memory.Record) ([]memory.QueryResult, error)
}

// EmbeddingIndex stores and searches vector embeddings.
type EmbeddingIndex interface {
	Upsert(ctx context.Context, recordID string, embedding []float32) error
	Search(ctx context.Context, embedding []float32, limit int) ([]string, error)
}
