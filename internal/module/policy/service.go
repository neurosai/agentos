package policymod

import (
	"context"
	"errors"
	"fmt"

	"github.com/neurosai/agentos/internal/adapter/idgen"
	"github.com/neurosai/agentos/internal/adapter/opa"
	auditmod "github.com/neurosai/agentos/internal/module/audit"
	"github.com/neurosai/agentos/internal/domain/policy"
	"github.com/neurosai/agentos/internal/port"
)

// Service wraps OPA evaluation and audit recording.
type Service struct {
	eval    opa.Evaluator
	audit   *auditmod.Service
	ids     port.IDGenerator
	tenant  string
	subject string
}

func NewService(eval opa.Evaluator, audit *auditmod.Service, tenant, subject string) *Service {
	return &Service{
		eval:    eval,
		audit:   audit,
		ids:     idgen.UUID{},
		tenant:  tenant,
		subject: subject,
	}
}

func (s *Service) Evaluate(ctx context.Context, in policy.EvaluationInput) (policy.Decision, error) {
	pkg := opa.PackageForAction(in.Action)
	dec, err := s.eval.Evaluate(ctx, pkg, in)
	if err != nil {
		return policy.Decision{}, err
	}
	id, err := s.ids.New("dec")
	if err != nil {
		return policy.Decision{}, err
	}
	dec.ID = id
	traceID := in.RequestID
	if traceID == "" {
		traceID, _ = s.ids.New("req")
	}
	_, _ = s.audit.Record(ctx, auditmod.RecordInput{
		TenantID:  s.tenant,
		SubjectID: s.subject,
		TaskID:    in.Context.TaskID,
		EventType: "policy.decided",
		Action:    in.Action,
		Decision:  string(dec.Effect),
		TraceID:   traceID,
		Payload: map[string]any{
			"effect":      dec.Effect,
			"denyReason":  dec.DenyReason,
			"resourceId":  in.Resource.ID,
			"resourceType": in.Resource.Type,
		},
	})
	return dec, nil
}

func (s *Service) Compile(ctx context.Context, in policy.CompileInput) (policy.CompiledFilters, error) {
	return policy.CompiledFilters{}, fmt.Errorf("compile not implemented in v0.2")
}

var ErrDenied = errors.New("policy denied")

func (s *Service) EnsureAllow(dec policy.Decision) error {
	if dec.Effect == policy.EffectDeny {
		msg := dec.DenyReason
		if msg == "" {
			msg = "policy denied"
		}
		return fmt.Errorf("%w: %s", ErrDenied, msg)
	}
	return nil
}

func IsDenied(err error) bool {
	return errors.Is(err, ErrDenied)
}

func RequiresApproval(dec policy.Decision) bool {
	return dec.Effect == policy.EffectRequireApproval
}
