package usecase

import (
	"context"

	"github.com/neurosai/agentos/internal/domain/memory"
)

// MemoryService governs memory read/write/query with policy prefiltering.
//
// Pre: policy evaluates memory.write/read; ACL enforced in substrate.
// Post: provenance recorded; audit events memory.write/query appended.
type MemoryService interface {
	Write(ctx context.Context, record memory.Record) (memory.Record, error)
	Get(ctx context.Context, id string) (memory.Record, error)
	Query(ctx context.Context, q memory.Query) ([]memory.QueryResult, error)
	Forget(ctx context.Context, id string) error
	Promote(ctx context.Context, id string, targetType memory.Type) (memory.Record, error)
	Revise(ctx context.Context, id string, revised memory.Record) (memory.Record, error)
}
