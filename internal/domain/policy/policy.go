package policy

import "time"

// Effect is the outcome of a policy evaluation.
type Effect string

const (
	EffectAllow           Effect = "allow"
	EffectDeny            Effect = "deny"
	EffectRequireApproval Effect = "require_approval"
	EffectRedact          Effect = "redact"
	EffectFilter          Effect = "filter"
	EffectExchangeToken   Effect = "exchange_token"
	EffectSandboxOnly     Effect = "sandbox_only"
)

// Subject identifies who is acting.
type Subject struct {
	Type   string
	ID     string
	Roles  []string
	Groups []string
}

// ActingAgent identifies the runtime agent principal.
type ActingAgent struct {
	ID     string
	Labels map[string]string
}

// Resource is the target of an action.
type Resource struct {
	Type      string
	ID        string
	TenantID  string
	Risk      string
	Namespace string
}

// Context carries ambient decision attributes.
type Context struct {
	TaskID         string
	Classification string
	Workspace      string
	SourceTrust    string
	Environment    string
}

// EvaluationInput is the normalized document sent to OPA.
type EvaluationInput struct {
	RequestID   string
	Subject     Subject
	ActingAgent ActingAgent
	Resource    Resource
	Action      string
	Context     Context
}

// Obligation is an action the enforcer must perform when allowing.
type Obligation struct {
	Type   string
	Params map[string]string
}

// Filter is a partial evaluation result for data queries.
type Filter struct {
	Field    string
	Operator string
	Value    string
}

// Decision is the policy engine outcome.
type Decision struct {
	ID          string
	Effect      Effect
	Obligations []Obligation
	Filters     []Filter
	DenyReason  string
	EvaluatedAt time.Time
}

// CompileInput requests partial evaluation for query prefiltering.
type CompileInput struct {
	Package string
	Query   string
	Input   EvaluationInput
}

// CompiledFilters holds compile API output for MemoryD/CatalogD.
type CompiledFilters struct {
	Filters []Filter
	Unknown []string
}

// Allows returns true when the decision permits the action.
func (d Decision) Allows() bool {
	return d.Effect == EffectAllow || d.Effect == EffectExchangeToken || d.Effect == EffectFilter
}
