package memorymod

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/neurosai/agentos/internal/domain/memory"
	"github.com/neurosai/agentos/internal/domain/policy"
	auditmod "github.com/neurosai/agentos/internal/module/audit"
	policymod "github.com/neurosai/agentos/internal/module/policy"
	"github.com/neurosai/agentos/internal/port"
	"github.com/neurosai/agentos/pkg/ids"
)

// Service implements governed memory operations.
type Service struct {
	repo    port.MemoryRepository
	policy  *policymod.Service
	audit   *auditmod.Service
	clock   port.Clock
	ids     port.IDGenerator
	tenant  string
	subject string
	roles   []string
	groups  []string
}

type Options struct {
	Repo    port.MemoryRepository
	Policy  *policymod.Service
	Audit   *auditmod.Service
	Clock   port.Clock
	IDs     port.IDGenerator
	Tenant  string
	Subject string
	Roles   []string
	Groups  []string
}

func NewService(opt Options) *Service {
	return &Service{
		repo:    opt.Repo,
		policy:  opt.Policy,
		audit:   opt.Audit,
		clock:   opt.Clock,
		ids:     opt.IDs,
		tenant:  opt.Tenant,
		subject: opt.Subject,
		roles:   opt.Roles,
		groups:  opt.Groups,
	}
}

func (s *Service) policySubject() policy.Subject {
	return policy.Subject{ID: s.subject, Roles: s.roles, Groups: s.groups}
}

func (s *Service) Create(ctx context.Context, rec memory.Record) (memory.Record, error) {
	dec, err := s.policy.Evaluate(ctx, policy.EvaluationInput{
		Subject:  s.policySubject(),
		Resource: policy.Resource{Type: "memory", TenantID: s.tenant, Namespace: rec.Namespace},
		Action:   "write",
		Context:  policy.Context{Classification: rec.Classification},
	})
	if err != nil {
		return memory.Record{}, err
	}
	if err := s.policy.EnsureAllow(dec); err != nil {
		return memory.Record{}, err
	}
	if rec.ID == "" {
		rec.ID = uuid.New().String()
	}
	if rec.TenantID == "" {
		rec.TenantID = s.tenant
	}
	if rec.CreatedBy == "" {
		rec.CreatedBy = s.subject
	}
	rec.CreatedAt = s.clock.Now().UTC()
	if err := s.repo.Create(ctx, rec); err != nil {
		return memory.Record{}, err
	}
	trace, _ := s.ids.New(string(ids.PrefixRequest))
	_, _ = s.audit.Record(ctx, auditmod.RecordInput{
		TenantID:     s.tenant,
		SubjectID:    s.subject,
		EventType:    "memory.write",
		ResourceType: "memory",
		ResourceID:   rec.ID,
		Action:       "memory.write",
		TraceID:      trace,
		Payload:      map[string]any{"namespace": rec.Namespace, "type": rec.Type},
	})
	return rec, nil
}

func (s *Service) Get(ctx context.Context, id string) (memory.Record, error) {
	rec, err := s.repo.Get(ctx, id)
	if err != nil {
		return memory.Record{}, err
	}
	if err := s.checkACL(rec); err != nil {
		return memory.Record{}, err
	}
	if rec.DeletedAt != nil {
		return memory.Record{}, fmt.Errorf("memory record deleted")
	}
	if rec.RetentionUntil != nil && s.clock.Now().After(*rec.RetentionUntil) {
		return memory.Record{}, fmt.Errorf("memory record expired")
	}
	dec, err := s.policy.Evaluate(ctx, policy.EvaluationInput{
		Subject:  s.policySubject(),
		Resource: policy.Resource{Type: "memory", ID: id, TenantID: s.tenant, Namespace: rec.Namespace},
		Action:   "read",
		Record: policy.RecordContext{
			Classification: rec.Classification,
			Namespace:      rec.Namespace,
		},
	})
	if err != nil {
		return memory.Record{}, err
	}
	if err := s.policy.EnsureAllow(dec); err != nil {
		return memory.Record{}, err
	}
	trace, _ := s.ids.New(string(ids.PrefixRequest))
	_, _ = s.audit.Record(ctx, auditmod.RecordInput{
		TenantID:     s.tenant,
		SubjectID:    s.subject,
		EventType:    "memory.read",
		ResourceType: "memory",
		ResourceID:   id,
		Action:       "memory.read",
		TraceID:      trace,
	})
	return rec, nil
}

func (s *Service) Search(ctx context.Context, q memory.Query) ([]memory.QueryResult, error) {
	dec, err := s.policy.Evaluate(ctx, policy.EvaluationInput{
		Subject:  s.policySubject(),
		Resource: policy.Resource{Type: "memory", TenantID: s.tenant, Namespace: q.Namespace},
		Action:   "search",
	})
	if err != nil {
		return nil, err
	}
	if err := s.policy.EnsureAllow(dec); err != nil {
		return nil, err
	}
	results, err := s.repo.Query(ctx, q, nil)
	if err != nil {
		return nil, err
	}
	now := s.clock.Now()
	filtered := make([]memory.QueryResult, 0, len(results))
	for _, r := range results {
		if r.Record.DeletedAt != nil {
			continue
		}
		if r.Record.RetentionUntil != nil && now.After(*r.Record.RetentionUntil) {
			continue
		}
		if err := s.checkACL(r.Record); err != nil {
			continue
		}
		filtered = append(filtered, r)
	}
	trace, _ := s.ids.New(string(ids.PrefixRequest))
	_, _ = s.audit.Record(ctx, auditmod.RecordInput{
		TenantID:  s.tenant,
		SubjectID: s.subject,
		EventType: "memory.read",
		Action:    "memory.search",
		TraceID:   trace,
		Payload:   map[string]any{"query": q.QueryText, "hits": len(filtered)},
	})
	return filtered, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	rec, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	dec, err := s.policy.Evaluate(ctx, policy.EvaluationInput{
		Subject:  s.policySubject(),
		Resource: policy.Resource{Type: "memory", ID: id, TenantID: s.tenant},
		Action:   "delete",
	})
	if err != nil {
		return err
	}
	if err := s.policy.EnsureAllow(dec); err != nil {
		return err
	}
	if err := s.repo.SoftDelete(ctx, id); err != nil {
		return err
	}
	trace, _ := s.ids.New(string(ids.PrefixRequest))
	_, _ = s.audit.Record(ctx, auditmod.RecordInput{
		TenantID:     s.tenant,
		SubjectID:    s.subject,
		EventType:    "memory.delete",
		ResourceType: "memory",
		ResourceID:   id,
		Action:       "memory.delete",
		TraceID:      trace,
		Payload:      map[string]any{"namespace": rec.Namespace},
	})
	return nil
}

func (s *Service) checkACL(rec memory.Record) error {
	if len(rec.ACL.View) == 0 {
		return nil
	}
	for _, v := range rec.ACL.View {
		if v == s.subject {
			return nil
		}
	}
	for _, r := range s.roles {
		if contains(rec.ACL.View, "role:"+r) {
			return nil
		}
	}
	for _, g := range s.groups {
		if contains(rec.ACL.View, g) || contains(rec.ACL.View, "group:"+strings.TrimPrefix(g, "group:")) {
			return nil
		}
	}
	return fmt.Errorf("acl denied")
}

func contains(slice []string, want string) bool {
	for _, v := range slice {
		if v == want {
			return true
		}
	}
	return false
}
