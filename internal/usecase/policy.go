package usecase

import (
	"context"

	"github.com/neurosai/agentos/internal/domain/policy"
)

// PolicyService exposes policy evaluation to northbound and internal callers.
//
// Pre: caller authenticated with policy.evaluate or internal service identity.
// Post: decision persisted for introspection; audit event policy.evaluate appended.
type PolicyService interface {
	Evaluate(ctx context.Context, pkg, query string, input policy.EvaluationInput) (policy.Decision, error)
	Compile(ctx context.Context, in policy.CompileInput) (policy.CompiledFilters, error)
	GetDecision(ctx context.Context, decisionID string) (policy.Decision, error)
	PutModule(ctx context.Context, name, source string) error
}
