package auditmod

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/neurosai/agentos/internal/adapter/idgen"
	"github.com/neurosai/agentos/internal/domain/audit"
	"github.com/neurosai/agentos/internal/port"
	"github.com/neurosai/agentos/pkg/ids"
)

// RecordInput carries fields for a new audit event.
type RecordInput struct {
	TenantID     string
	SubjectID    string
	AgentID      string
	TaskID       string
	EventType    string
	ResourceType string
	ResourceID   string
	Action       string
	Decision     string
	Status       string
	TraceID      string
	SpanID       string
	Payload      map[string]any
}

// Service implements audit operations with hash chaining.
type Service struct {
	repo  port.AuditRepository
	clock port.Clock
	ids   port.IDGenerator
}

func NewService(repo port.AuditRepository, clock port.Clock, ids port.IDGenerator) *Service {
	if clock == nil {
		panic("clock required")
	}
	if ids == nil {
		ids = idgen.UUID{}
	}
	return &Service{repo: repo, clock: clock, ids: ids}
}

func (s *Service) Record(ctx context.Context, in RecordInput) (audit.Event, error) {
	stream := in.TenantID
	if stream == "" {
		stream = "default"
	}
	prev, err := s.repo.LastHash(ctx, stream)
	if err != nil {
		return audit.Event{}, err
	}
	payloadHash, err := hashPayload(in.Payload)
	if err != nil {
		return audit.Event{}, err
	}
	eventID, err := s.ids.New(string(ids.PrefixEvent))
	if err != nil {
		return audit.Event{}, err
	}
	occurred := s.clock.Now().UTC()
	eventHash := chainHash(prev, payloadHash, eventID, occurred)
	ev := audit.Event{
		EventID:      eventID,
		OccurredAt:   occurred,
		TenantID:     in.TenantID,
		SubjectID:    in.SubjectID,
		AgentID:      in.AgentID,
		TaskID:       in.TaskID,
		EventType:    in.EventType,
		ResourceType: in.ResourceType,
		ResourceID:   in.ResourceID,
		Action:       in.Action,
		Decision:     in.Decision,
		Status:       in.Status,
		PayloadHash:  payloadHash,
		PrevHash:     prev,
		EventHash:    eventHash,
		TraceID:      in.TraceID,
		SpanID:       in.SpanID,
	}
	if err := s.repo.Insert(ctx, ev); err != nil {
		return audit.Event{}, err
	}
	return ev, nil
}

func (s *Service) Query(ctx context.Context, tenantID, taskID string, limit int) ([]audit.Event, error) {
	return s.repo.List(ctx, port.AuditQueryFilter{
		TenantID: tenantID,
		TaskID:   taskID,
		Limit:    limit,
	})
}

func (s *Service) QueryByTrace(ctx context.Context, traceID string) ([]audit.Event, error) {
	type traceLister interface {
		ListByTrace(ctx context.Context, traceID string) ([]audit.Event, error)
	}
	if tl, ok := s.repo.(traceLister); ok {
		return tl.ListByTrace(ctx, traceID)
	}
	events, err := s.repo.List(ctx, port.AuditQueryFilter{Limit: 1000})
	if err != nil {
		return nil, err
	}
	var out []audit.Event
	for _, e := range events {
		if e.TraceID == traceID {
			out = append(out, e)
		}
	}
	return out, nil
}

func (s *Service) Get(ctx context.Context, eventID string) (audit.Event, error) {
	return s.repo.GetByID(ctx, eventID)
}

func (s *Service) Proof(ctx context.Context, streamID string) (audit.Proof, error) {
	events, err := s.repo.List(ctx, port.AuditQueryFilter{TenantID: streamID, Limit: 10000})
	if err != nil {
		return audit.Proof{}, err
	}
	if len(events) == 0 {
		return audit.Proof{}, fmt.Errorf("empty stream")
	}
	sort.Slice(events, func(i, j int) bool {
		return events[i].OccurredAt.Before(events[j].OccurredAt)
	})
	root := events[len(events)-1].EventHash
	if err := verifyChain(events); err != nil {
		return audit.Proof{}, err
	}
	return audit.Proof{
		StreamID:   streamID,
		FromEvent:  events[0].EventID,
		ToEvent:    events[len(events)-1].EventID,
		RootHash:   root,
		VerifiedAt: s.clock.Now().UTC(),
	}, nil
}

func hashPayload(payload map[string]any) (string, error) {
	if payload == nil {
		payload = map[string]any{}
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

func chainHash(prev, payloadHash, eventID string, at time.Time) string {
	h := sha256.New()
	_, _ = h.Write([]byte(prev))
	_, _ = h.Write([]byte(payloadHash))
	_, _ = h.Write([]byte(eventID))
	_, _ = h.Write([]byte(at.Format(time.RFC3339Nano)))
	return hex.EncodeToString(h.Sum(nil))
}

func verifyChain(events []audit.Event) error {
	var prev string
	for _, e := range events {
		if e.PrevHash != prev && !(prev == "" && e.PrevHash == "") {
			return fmt.Errorf("chain break at %s", e.EventID)
		}
		expected := chainHash(e.PrevHash, e.PayloadHash, e.EventID, e.OccurredAt)
		if e.EventHash != expected {
			return fmt.Errorf("invalid hash at %s", e.EventID)
		}
		prev = e.EventHash
	}
	return nil
}

// StreamIDFromTenant returns audit stream key.
func StreamIDFromTenant(tenantID string) string {
	if strings.TrimSpace(tenantID) == "" {
		return "default"
	}
	return tenantID
}
