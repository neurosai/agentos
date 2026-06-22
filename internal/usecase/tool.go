package usecase

import (
	"context"

	"github.com/neurosai/agentos/internal/domain/tool"
)

// ToolService is the syscall boundary for MCP and built-in tools.
//
// Pre: policy preflight on every invoke; credentials resolved at call time.
// Post: invocation recorded; audit event tool.invoke appended.
type ToolService interface {
	List(ctx context.Context, tenantID string) ([]tool.Ref, error)
	Get(ctx context.Context, toolID string) (tool.Ref, error)
	Invoke(ctx context.Context, req tool.Invoke) (tool.Result, error)
	DryRun(ctx context.Context, req tool.Invoke) (tool.Result, error)
	Schema(ctx context.Context, toolID string) (map[string]any, error)
}
