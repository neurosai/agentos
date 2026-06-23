package toolmod

import (
	"context"
	"fmt"
	"sync"

	"github.com/neurosai/agentos/internal/adapter/builtin"
	"github.com/neurosai/agentos/internal/domain/policy"
	"github.com/neurosai/agentos/internal/domain/task"
	"github.com/neurosai/agentos/internal/domain/tool"
	auditmod "github.com/neurosai/agentos/internal/module/audit"
	policymod "github.com/neurosai/agentos/internal/module/policy"
	taskmod "github.com/neurosai/agentos/internal/module/task"
	"github.com/neurosai/agentos/internal/port"
	"github.com/neurosai/agentos/pkg/ids"
)

// PendingInvoke stores approval-gated tool calls.
type PendingInvoke struct {
	Request tool.Invoke
	Ref     tool.Ref
}

// Service implements ToolD syscall boundary.
type Service struct {
	registry   port.ToolRegistry
	invoker    *builtin.Registry
	invocations port.ToolInvocationRepository
	policy     *policymod.Service
	audit      *auditmod.Service
	tasks      *taskmod.Service
	clock      port.Clock
	ids        port.IDGenerator
	tenant     string
	subject    string
	roles      []string

	mu       sync.Mutex
	pending  map[string]PendingInvoke
}

type Options struct {
	Registry    port.ToolRegistry
	Invoker     *builtin.Registry
	Invocations port.ToolInvocationRepository
	Policy      *policymod.Service
	Audit       *auditmod.Service
	Tasks       *taskmod.Service
	Clock       port.Clock
	IDs         port.IDGenerator
	Tenant      string
	Subject     string
	Roles       []string
}

func NewService(opt Options) *Service {
	return &Service{
		registry:    opt.Registry,
		invoker:     opt.Invoker,
		invocations: opt.Invocations,
		policy:      opt.Policy,
		audit:       opt.Audit,
		tasks:       opt.Tasks,
		clock:       opt.Clock,
		ids:         opt.IDs,
		tenant:      opt.Tenant,
		subject:     opt.Subject,
		roles:       opt.Roles,
		pending:     make(map[string]PendingInvoke),
	}
}

func (s *Service) List(ctx context.Context, tenantID string) ([]tool.Ref, error) {
	return s.registry.List(ctx, tenantID)
}

func (s *Service) Get(ctx context.Context, toolID string) (tool.Ref, error) {
	return s.registry.Get(ctx, toolID)
}

func (s *Service) Invoke(ctx context.Context, req tool.Invoke) (tool.Result, error) {
	if req.IdempotencyKey != "" && s.invocations != nil {
		existing, err := s.invocations.GetByIdempotency(ctx, req.ToolID, req.IdempotencyKey)
		if err == nil && existing.Status == "completed" {
			return tool.Result{
				CallID:       existing.ID,
				Status:       existing.Status,
				Output:       map[string]any{"idempotent": true},
				AuditEventID: existing.AuditEventID,
				CompletedAt:  s.clock.Now().UTC(),
			}, nil
		}
	}
	ref, err := s.registry.Get(ctx, req.ToolID)
	if err != nil {
		return tool.Result{}, err
	}
	trace, _ := s.ids.New(string(ids.PrefixRequest))
	dec, err := s.policy.Evaluate(ctx, policy.EvaluationInput{
		RequestID: trace,
		Subject:   policy.Subject{ID: s.subject, Roles: s.roles},
		ActingAgent: policy.ActingAgent{ID: req.AgentID},
		Resource: policy.Resource{
			Type:     "tool",
			ID:       ref.ID,
			TenantID: s.tenant,
			Risk:     ref.Risk,
		},
		Action: "invoke",
		Context: policy.Context{
			TaskID:         req.TaskID,
			Classification: req.Context.Classification,
			SourceTrust:    req.Context.SourceTrust,
			Workspace:      req.Context.Workspace,
		},
	})
	if err != nil {
		return tool.Result{}, err
	}
	if dec.Effect == policy.EffectDeny {
		return tool.Result{}, fmt.Errorf("policy denied: %s", dec.DenyReason)
	}
	if policymod.RequiresApproval(dec) {
		approvalID, err := s.ids.New(string(ids.PrefixEvent))
		if err != nil {
			return tool.Result{}, err
		}
		s.mu.Lock()
		s.pending[req.TaskID] = PendingInvoke{Request: req, Ref: ref}
		s.mu.Unlock()
		if err := s.tasks.RequestApproval(ctx, req.TaskID, task.Approval{ID: approvalID}); err != nil {
			return tool.Result{}, err
		}
		return tool.Result{
			Status:      "waiting_approval",
			CompletedAt: s.clock.Now().UTC(),
		}, nil
	}
	return s.execute(ctx, req, ref, trace)
}

func (s *Service) ResumeAfterApproval(ctx context.Context, taskID string) (tool.Result, error) {
	s.mu.Lock()
	p, ok := s.pending[taskID]
	delete(s.pending, taskID)
	s.mu.Unlock()
	if !ok {
		return tool.Result{}, fmt.Errorf("no pending invoke for task %s", taskID)
	}
	trace, _ := s.ids.New(string(ids.PrefixRequest))
	return s.execute(ctx, p.Request, p.Ref, trace)
}

func (s *Service) execute(ctx context.Context, req tool.Invoke, ref tool.Ref, trace string) (tool.Result, error) {
	callID, err := s.ids.New(string(ids.PrefixToolCall))
	if err != nil {
		return tool.Result{}, err
	}
	out, err := s.invoker.Invoke(ctx, req.ToolID, req.Arguments)
	if err != nil {
		return tool.Result{}, err
	}
	ev, _ := s.audit.Record(ctx, auditmod.RecordInput{
		TenantID:     s.tenant,
		SubjectID:    s.subject,
		AgentID:      req.AgentID,
		TaskID:       req.TaskID,
		EventType:    "tool.invoke",
		ResourceType: "tool",
		ResourceID:   ref.ID,
		Action:       "tool.invoke",
		TraceID:      trace,
		Payload:      map[string]any{"output": out, "toolId": ref.ID},
	})
	now := s.clock.Now().UTC()
	completed := now
	rec := tool.InvocationRecord{
		ID:             callID,
		ToolID:         ref.ID,
		TaskID:         req.TaskID,
		AgentID:        req.AgentID,
		Status:         "completed",
		AuditEventID:   ev.EventID,
		IdempotencyKey: req.IdempotencyKey,
		StartedAt:      now,
		CompletedAt:    &completed,
	}
	if s.invocations != nil {
		_ = s.invocations.Create(ctx, rec)
	}
	return tool.Result{
		CallID:       callID,
		Status:       "completed",
		Output:       out,
		OutputTrust:  "system",
		AuditEventID: ev.EventID,
		CompletedAt:  now,
	}, nil
}

func (s *Service) DryRun(ctx context.Context, req tool.Invoke) (tool.Result, error) {
	ref, err := s.registry.Get(ctx, req.ToolID)
	if err != nil {
		return tool.Result{}, err
	}
	return tool.Result{
		Status:      "dry_run",
		Output:      map[string]any{"tool": ref.ID, "risk": ref.Risk},
		CompletedAt: s.clock.Now().UTC(),
	}, nil
}

func (s *Service) Schema(ctx context.Context, toolID string) (map[string]any, error) {
	ref, err := s.registry.Get(ctx, toolID)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"id":          ref.ID,
		"name":        ref.Name,
		"transport":   ref.Transport,
		"risk":        ref.Risk,
		"description": ref.Description,
	}, nil
}
