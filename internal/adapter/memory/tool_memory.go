package memory

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/neurosai/agentos/internal/domain/memory"
	"github.com/neurosai/agentos/internal/domain/tool"
	"github.com/neurosai/agentos/internal/port"
)

// MemoryRepository is an in-memory memory store.
type MemoryRepository struct {
	mu      sync.RWMutex
	records map[string]memory.Record
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{records: make(map[string]memory.Record)}
}

func (r *MemoryRepository) Create(ctx context.Context, rec memory.Record) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.records[rec.ID] = rec
	return nil
}

func (r *MemoryRepository) Get(ctx context.Context, id string) (memory.Record, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rec, ok := r.records[id]
	if !ok {
		return memory.Record{}, errNotFound("memory", id)
	}
	return rec, nil
}

func (r *MemoryRepository) SoftDelete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	rec, ok := r.records[id]
	if !ok {
		return errNotFound("memory", id)
	}
	now := time.Now().UTC()
	rec.DeletedAt = &now
	r.records[id] = rec
	return nil
}

func (r *MemoryRepository) Revise(ctx context.Context, id string, revised memory.Record) error {
	return r.Create(ctx, revised)
}

func (r *MemoryRepository) Query(ctx context.Context, q memory.Query, _ []memory.Record) ([]memory.QueryResult, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []memory.QueryResult
	for _, rec := range r.records {
		if rec.DeletedAt != nil {
			continue
		}
		if q.Namespace != "" && rec.Namespace != q.Namespace {
			continue
		}
		if q.QueryText != "" && !strings.Contains(strings.ToLower(rec.Content), strings.ToLower(q.QueryText)) {
			continue
		}
		out = append(out, memory.QueryResult{Record: rec, Score: 1})
		if q.Limit > 0 && len(out) >= q.Limit {
			break
		}
	}
	return out, nil
}

var _ port.MemoryRepository = (*MemoryRepository)(nil)

// ToolInvocationRepository stores invocations in memory.
type ToolInvocationRepository struct {
	mu    sync.RWMutex
	byID  map[string]tool.InvocationRecord
	byKey map[string]tool.InvocationRecord
}

func NewToolInvocationRepository() *ToolInvocationRepository {
	return &ToolInvocationRepository{
		byID:  make(map[string]tool.InvocationRecord),
		byKey: make(map[string]tool.InvocationRecord),
	}
}

func (r *ToolInvocationRepository) Create(ctx context.Context, rec tool.InvocationRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID[rec.ID] = rec
	if rec.IdempotencyKey != "" {
		r.byKey[rec.ToolID+":"+rec.IdempotencyKey] = rec
	}
	return nil
}

func (r *ToolInvocationRepository) Get(ctx context.Context, callID string) (tool.InvocationRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rec, ok := r.byID[callID]
	if !ok {
		return tool.InvocationRecord{}, errNotFound("invocation", callID)
	}
	return rec, nil
}

func (r *ToolInvocationRepository) GetByIdempotency(ctx context.Context, toolID, key string) (tool.InvocationRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rec, ok := r.byKey[toolID+":"+key]
	if !ok {
		return tool.InvocationRecord{}, errNotFound("invocation", key)
	}
	return rec, nil
}

var _ port.ToolInvocationRepository = (*ToolInvocationRepository)(nil)
