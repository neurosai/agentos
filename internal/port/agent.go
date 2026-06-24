package port

import (
	"context"

	"github.com/neurosai/agentos/internal/domain/agent"
)

// AgentRuntime executes an agent against a task.
type AgentRuntime interface {
	Run(ctx context.Context, spec agent.RunSpec) error
}
