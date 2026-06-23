package memory

import (
	"context"
	"sync"

	"github.com/neurosai/agentos/internal/domain/audit"
	"github.com/neurosai/agentos/internal/port"
)

// AuditRepository is an in-memory audit store with hash chain support.
type AuditRepository struct {
	mu     sync.RWMutex
	events map[string]audit.Event
	order  []string
	byTask map[string][]string
	byTrace map[string][]string
	last   map[string]string
}

func NewAuditRepository() *AuditRepository {
	return &AuditRepository{
		events:  make(map[string]audit.Event),
		byTask:  make(map[string][]string),
		byTrace: make(map[string][]string),
		last:    make(map[string]string),
	}
}

func (r *AuditRepository) Insert(ctx context.Context, event audit.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events[event.EventID] = event
	r.order = append(r.order, event.EventID)
	if event.TaskID != "" {
		r.byTask[event.TaskID] = append(r.byTask[event.TaskID], event.EventID)
	}
	if event.TraceID != "" {
		r.byTrace[event.TraceID] = append(r.byTrace[event.TraceID], event.EventID)
	}
	stream := event.TenantID
	if stream == "" {
		stream = "default"
	}
	r.last[stream] = event.EventHash
	return nil
}

func (r *AuditRepository) GetByID(ctx context.Context, eventID string) (audit.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	e, ok := r.events[eventID]
	if !ok {
		return audit.Event{}, errNotFound("audit event", eventID)
	}
	return e, nil
}

func (r *AuditRepository) List(ctx context.Context, filter port.AuditQueryFilter) ([]audit.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var ids []string
	switch {
	case filter.TaskID != "":
		ids = r.byTask[filter.TaskID]
	case filter.TenantID != "":
		for _, id := range r.order {
			if r.events[id].TenantID == filter.TenantID {
				ids = append(ids, id)
			}
		}
	default:
		ids = append(ids, r.order...)
	}
	limit := filter.Limit
	if limit <= 0 {
		limit = 100
	}
	if len(ids) > limit {
		ids = ids[len(ids)-limit:]
	}
	out := make([]audit.Event, 0, len(ids))
	for _, id := range ids {
		out = append(out, r.events[id])
	}
	return out, nil
}

func (r *AuditRepository) LastHash(ctx context.Context, streamID string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.last[streamID], nil
}

func (r *AuditRepository) ListByTrace(ctx context.Context, traceID string) ([]audit.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids := r.byTrace[traceID]
	out := make([]audit.Event, 0, len(ids))
	for _, id := range ids {
		out = append(out, r.events[id])
	}
	return out, nil
}

var _ port.AuditRepository = (*AuditRepository)(nil)
