package port

import (
	"context"

	"github.com/neurosai/agentos/internal/domain/policy"
)

// PolicyEvaluator evaluates and compiles Rego policies.
type PolicyEvaluator interface {
	Evaluate(ctx context.Context, in policy.EvaluationInput) (policy.Decision, error)
	Compile(ctx context.Context, in policy.CompileInput) (policy.CompiledFilters, error)
}

// PolicyModuleStore loads Rego bundles.
type PolicyModuleStore interface {
	PutModule(ctx context.Context, name, source string) error
	GetModule(ctx context.Context, name string) (string, error)
}
