package port

import (
	"context"

	"github.com/neurosai/agentos/internal/domain/tool"
)

// ToolRegistry lists registered tools.
type ToolRegistry interface {
	List(ctx context.Context, tenantID string) ([]tool.Ref, error)
	Get(ctx context.Context, toolID string) (tool.Ref, error)
	Upsert(ctx context.Context, ref tool.Ref) error
}

// ToolInvoker executes policy-bound tool calls.
type ToolInvoker interface {
	Invoke(ctx context.Context, req tool.Invoke) (tool.Result, error)
	DryRun(ctx context.Context, req tool.Invoke) (tool.Result, error)
}

// ToolInvocationRepository persists invocation audit cross-refs.
type ToolInvocationRepository interface {
	Create(ctx context.Context, rec tool.InvocationRecord) error
	Get(ctx context.Context, callID string) (tool.InvocationRecord, error)
	GetByIdempotency(ctx context.Context, toolID, key string) (tool.InvocationRecord, error)
}
