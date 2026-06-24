package taskmod

import (
	"context"
	"fmt"

	mem "github.com/neurosai/agentos/internal/adapter/memory"
	"github.com/neurosai/agentos/internal/domain/policy"
	"github.com/neurosai/agentos/internal/domain/task"
	auditmod "github.com/neurosai/agentos/internal/module/audit"
	policymod "github.com/neurosai/agentos/internal/module/policy"
	"github.com/neurosai/agentos/internal/port"
	"github.com/neurosai/agentos/internal/usecase"
	"github.com/neurosai/agentos/pkg/ids"
)

// Service implements task lifecycle operations.
type Service struct {
	tasks      port.TaskRepository
	messages   port.TaskMessageRepository
	artifacts  port.TaskArtifactRepository
	approvals  port.TaskApprovalRepository
	policy     *policymod.Service
	audit      *auditmod.Service
	clock      port.Clock
	ids        port.IDGenerator
	tenant     string
	subject    string
	roles      []string
	broadcast  *mem.EventBroadcaster
	traceID    func() string
}

type Options struct {
	Tasks     port.TaskRepository
	Messages  port.TaskMessageRepository
	Artifacts port.TaskArtifactRepository
	Approvals port.TaskApprovalRepository
	Policy    *policymod.Service
	Audit     *auditmod.Service
	Clock     port.Clock
	IDs       port.IDGenerator
	Tenant    string
	Subject   string
	Roles     []string
	Broadcast *mem.EventBroadcaster
}

func NewService(opt Options) *Service {
	return &Service{
		tasks:     opt.Tasks,
		messages:  opt.Messages,
		artifacts: opt.Artifacts,
		approvals: opt.Approvals,
		policy:    opt.Policy,
		audit:     opt.Audit,
		clock:     opt.Clock,
		ids:       opt.IDs,
		tenant:    opt.Tenant,
		subject:   opt.Subject,
		roles:     opt.Roles,
		broadcast: opt.Broadcast,
		traceID: func() string {
			id, _ := opt.IDs.New(string(ids.PrefixRequest))
			return id
		},
	}
}

func (s *Service) Create(ctx context.Context, in usecase.CreateTaskInput) (task.Task, error) {
	trace := s.traceID()
	dec, err := s.policy.Evaluate(ctx, policy.EvaluationInput{
		RequestID: trace,
		Subject: policy.Subject{
			ID:    s.subject,
			Roles: s.roles,
		},
		Resource: policy.Resource{Type: "task", TenantID: s.tenant},
		Action:   "task.submit",
		Context:  policy.Context{Classification: in.Labels["classification"]},
	})
	if err != nil {
		return task.Task{}, err
	}
	if err := s.policy.EnsureAllow(dec); err != nil {
		return task.Task{}, err
	}
	tenant := in.TenantID
	if tenant == "" {
		tenant = s.tenant
	}
	taskID, err := s.ids.New(string(ids.PrefixTask))
	if err != nil {
		return task.Task{}, err
	}
	now := s.clock.Now().UTC()
	t := task.Task{
		ID:        taskID,
		TenantID:  tenant,
		ContextID: in.ContextID,
		AgentRef:  in.AgentRef,
		Status:    task.StatusAccepted,
		Input:     in.Input,
		Labels:    in.Labels,
		CreatedBy: in.CreatedBy,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if t.CreatedBy == "" {
		t.CreatedBy = s.subject
	}
	if err := s.tasks.Create(ctx, t); err != nil {
		return task.Task{}, err
	}
	_, _ = s.audit.Record(ctx, auditmod.RecordInput{
		TenantID:  tenant,
		SubjectID: s.subject,
		TaskID:    taskID,
		EventType: "task.created",
		Action:    "task.submit",
		TraceID:   trace,
		Payload: map[string]any{
			"agentRef": in.AgentRef,
			"status":   t.Status,
		},
	})
	s.emitStatus(taskID, string(t.Status))
	return t, nil
}

func (s *Service) Get(ctx context.Context, taskID string) (task.Task, error) {
	dec, err := s.policy.Evaluate(ctx, policy.EvaluationInput{
		Subject:  policy.Subject{ID: s.subject, Roles: s.roles},
		Resource: policy.Resource{Type: "task", ID: taskID, TenantID: s.tenant},
		Action:   "task.read",
		Context:  policy.Context{TaskID: taskID},
	})
	if err != nil {
		return task.Task{}, err
	}
	if err := s.policy.EnsureAllow(dec); err != nil {
		return task.Task{}, err
	}
	return s.tasks.Get(ctx, taskID)
}

func (s *Service) Cancel(ctx context.Context, taskID string) (task.Task, error) {
	dec, err := s.policy.Evaluate(ctx, policy.EvaluationInput{
		Subject:  policy.Subject{ID: s.subject, Roles: s.roles},
		Resource: policy.Resource{Type: "task", ID: taskID, TenantID: s.tenant},
		Action:   "task.cancel",
		Context:  policy.Context{TaskID: taskID},
	})
	if err != nil {
		return task.Task{}, err
	}
	if err := s.policy.EnsureAllow(dec); err != nil {
		return task.Task{}, err
	}
	t, err := s.tasks.Get(ctx, taskID)
	if err != nil {
		return task.Task{}, err
	}
	if err := task.Transition(&t, task.StatusCancelled); err != nil {
		return task.Task{}, err
	}
	t.UpdatedAt = s.clock.Now().UTC()
	if err := s.tasks.UpdateStatus(ctx, taskID, t.Status); err != nil {
		return task.Task{}, err
	}
	s.emitStatus(taskID, string(t.Status))
	return t, nil
}

func (s *Service) AddMessage(ctx context.Context, taskID string, msg task.Message) error {
	if msg.ID == "" {
		id, err := s.ids.New(string(ids.PrefixEvent))
		if err != nil {
			return err
		}
		msg.ID = id
	}
	msg.TaskID = taskID
	msg.CreatedAt = s.clock.Now().UTC()
	if err := s.messages.Append(ctx, msg); err != nil {
		return err
	}
	if s.broadcast != nil {
		s.broadcast.Publish(msg)
	}
	return nil
}

func (s *Service) RequestApproval(ctx context.Context, taskID string, approval task.Approval) error {
	if approval.ID == "" {
		id, err := s.ids.New(string(ids.PrefixEvent))
		if err != nil {
			return err
		}
		approval.ID = id
	}
	approval.TaskID = taskID
	approval.Requested = s.clock.Now().UTC()
	t, err := s.tasks.Get(ctx, taskID)
	if err != nil {
		return err
	}
	if err := task.Transition(&t, task.StatusWaitingApproval); err != nil {
		return err
	}
	if err := s.tasks.UpdateStatus(ctx, taskID, t.Status); err != nil {
		return err
	}
	if err := s.approvals.Create(ctx, approval); err != nil {
		return err
	}
	s.emitStatus(taskID, string(t.Status))
	return nil
}

func (s *Service) DecideApproval(ctx context.Context, taskID, approvalID string, approved bool, decidedBy string) (task.Task, error) {
	dec, err := s.policy.Evaluate(ctx, policy.EvaluationInput{
		Subject:  policy.Subject{ID: s.subject, Roles: s.roles},
		Resource: policy.Resource{Type: "task", ID: taskID, TenantID: s.tenant},
		Action:   "task.approve",
		Context:  policy.Context{TaskID: taskID},
	})
	if err != nil {
		return task.Task{}, err
	}
	if err := s.policy.EnsureAllow(dec); err != nil {
		return task.Task{}, err
	}
	if approvalID == "" {
		pending, err := s.approvals.GetPending(ctx, taskID)
		if err != nil {
			return task.Task{}, err
		}
		approvalID = pending.ID
	}
	if err := s.approvals.Decide(ctx, approvalID, approved, decidedBy); err != nil {
		return task.Task{}, err
	}
	t, err := s.tasks.Get(ctx, taskID)
	if err != nil {
		return task.Task{}, err
	}
	if approved {
		if err := task.Transition(&t, task.StatusRunning); err != nil {
			return task.Task{}, err
		}
	} else {
		if err := task.Transition(&t, task.StatusFailed); err != nil {
			return task.Task{}, err
		}
	}
	t.UpdatedAt = s.clock.Now().UTC()
	if err := s.tasks.UpdateStatus(ctx, taskID, t.Status); err != nil {
		return task.Task{}, err
	}
	s.emitStatus(taskID, string(t.Status))
	return t, nil
}

func (s *Service) UpdateStatus(ctx context.Context, taskID string, status task.Status) error {
	t, err := s.tasks.Get(ctx, taskID)
	if err != nil {
		return err
	}
	if err := task.Transition(&t, status); err != nil {
		return err
	}
	t.UpdatedAt = s.clock.Now().UTC()
	if err := s.tasks.UpdateStatus(ctx, taskID, t.Status); err != nil {
		return err
	}
	s.emitStatus(taskID, string(t.Status))
	return nil
}

func (s *Service) ListArtifacts(ctx context.Context, taskID string) ([]task.Artifact, error) {
	return s.artifacts.List(ctx, taskID)
}

func (s *Service) AddArtifact(ctx context.Context, a task.Artifact) error {
	if a.ID == "" {
		id, err := s.ids.New(string(ids.PrefixEvent))
		if err != nil {
			return err
		}
		a.ID = id
	}
	a.CreatedAt = s.clock.Now().UTC()
	return s.artifacts.Create(ctx, a)
}

func (s *Service) StreamEvents(ctx context.Context, taskID string, fromBeginning bool) (<-chan task.Message, error) {
	if s.broadcast == nil {
		ch := make(chan task.Message)
		close(ch)
		return ch, nil
	}
	if fromBeginning {
		msgs, _ := s.messages.List(ctx, taskID)
		ch := make(chan task.Message, len(msgs)+8)
		go func() {
			defer close(ch)
			for _, m := range msgs {
				select {
				case ch <- m:
				case <-ctx.Done():
					return
				}
			}
			sub := s.broadcast.Subscribe(taskID)
			defer func() {
				// drain subscription — simplified for v0.2
			}()
			for {
				select {
				case m, ok := <-sub:
					if !ok {
						return
					}
					ch <- m
				case <-ctx.Done():
					return
				}
			}
		}()
		return ch, nil
	}
	return s.broadcast.Subscribe(taskID), nil
}

func (s *Service) emitStatus(taskID, status string) {
	_ = s.AddMessage(context.Background(), taskID, task.Message{
		Role:    "system",
		Content: fmt.Sprintf(`{"type":"status","status":%q}`, status),
	})
}
