package opa

import (
	"context"
	"time"

	"github.com/neurosai/agentos/internal/domain/policy"
)

// Evaluator evaluates policy decisions.
type Evaluator interface {
	Evaluate(ctx context.Context, pkg string, input policy.EvaluationInput) (policy.Decision, error)
}

// AllowAll permits actions when OPA is not configured (local dev only).
type AllowAll struct{}

func (AllowAll) Evaluate(ctx context.Context, pkg string, input policy.EvaluationInput) (policy.Decision, error) {
	_ = ctx
	_ = pkg
	_ = input
	return policy.Decision{
		Effect:      policy.EffectAllow,
		EvaluatedAt: time.Now().UTC(),
	}, nil
}

// RequireHighRisk denies low paths but requires approval for high risk tools in stub mode.
type RequireHighRisk struct{}

func (RequireHighRisk) Evaluate(ctx context.Context, pkg string, input policy.EvaluationInput) (policy.Decision, error) {
	_ = ctx
	_ = pkg
	if input.Action == "invoke" && input.Resource.Risk == "high" {
		return policy.Decision{Effect: policy.EffectRequireApproval, EvaluatedAt: time.Now().UTC()}, nil
	}
	if input.Action == "invoke" && input.Resource.Risk != "" {
		return policy.Decision{Effect: policy.EffectAllow, EvaluatedAt: time.Now().UTC()}, nil
	}
	return policy.Decision{Effect: policy.EffectAllow, EvaluatedAt: time.Now().UTC()}, nil
}
