package agentmod

import (
	"context"
	"fmt"

	"github.com/neurosai/agentos/internal/domain/agent"
	"github.com/neurosai/agentos/internal/domain/memory"
	"github.com/neurosai/agentos/internal/domain/task"
	"github.com/neurosai/agentos/internal/domain/tool"
	auditmod "github.com/neurosai/agentos/internal/module/audit"
	memorymod "github.com/neurosai/agentos/internal/module/memory"
	taskmod "github.com/neurosai/agentos/internal/module/task"
	toolmod "github.com/neurosai/agentos/internal/module/tool"
	"github.com/neurosai/agentos/internal/port"
	"github.com/neurosai/agentos/pkg/ids"
)

// MockRuntime runs a fixed governed execution plan without an external agent process.
type MockRuntime struct {
	tasks  *taskmod.Service
	tools  *toolmod.Service
	memory *memorymod.Service
	audit  *auditmod.Service
	ids    port.IDGenerator
	tenant string
}

type MockOptions struct {
	Tasks  *taskmod.Service
	Tools  *toolmod.Service
	Memory *memorymod.Service
	Audit  *auditmod.Service
	IDs    port.IDGenerator
	Tenant string
}

func NewMockRuntime(opt MockOptions) *MockRuntime {
	return &MockRuntime{
		tasks:  opt.Tasks,
		tools:  opt.Tools,
		memory: opt.Memory,
		audit:  opt.Audit,
		ids:    opt.IDs,
		tenant: opt.Tenant,
	}
}

func (r *MockRuntime) Run(ctx context.Context, spec agent.RunSpec) error {
	manifest := spec.Manifest
	if spec.ManifestPath != "" && manifest.Name == "" {
		var err error
		manifest, err = LoadManifest(spec.ManifestPath)
		if err != nil {
			return err
		}
	}
	agentID := manifest.AgentID
	if agentID == "" {
		agentID = spec.AgentRef
	}

	t, err := r.tasks.Get(ctx, spec.TaskID)
	if err != nil {
		return err
	}
	if err := r.tasks.UpdateStatus(ctx, spec.TaskID, task.StatusRunning); err != nil {
		return err
	}

	goal, _ := t.Input["goal"].(string)
	if goal == "" {
		goal = fmt.Sprintf("task %s", spec.TaskID)
	}

	_, err = r.tools.Invoke(ctx, tool.Invoke{
		TaskID:    spec.TaskID,
		AgentID:   agentID,
		ToolID:    "tool.echo",
		Arguments: map[string]any{"message": goal},
	})
	if err != nil {
		_ = r.tasks.UpdateStatus(ctx, spec.TaskID, task.StatusFailed)
		return fmt.Errorf("tool echo: %w", err)
	}

	_, err = r.memory.Create(ctx, memory.Record{
		TenantID:       r.tenant,
		Namespace:      "workspace:payments",
		Type:           memory.TypeEvidence,
		Content:        fmt.Sprintf("agent %s completed goal: %s", agentID, goal),
		Classification: "internal",
		Confidence:     0.9,
		SourceRef:      "task:" + spec.TaskID,
		SourceType:     "agent_run",
	})
	if err != nil {
		_ = r.tasks.UpdateStatus(ctx, spec.TaskID, task.StatusFailed)
		return fmt.Errorf("memory write: %w", err)
	}

	artID, err := r.ids.New(string(ids.PrefixEvent))
	if err != nil {
		return err
	}
	if err := r.tasks.AddArtifact(ctx, task.Artifact{
		ID:          artID,
		TaskID:      spec.TaskID,
		Name:        "run-summary",
		ContentType: "application/json",
		URI:         "artifact://" + spec.TaskID + "/summary",
	}); err != nil {
		_ = r.tasks.UpdateStatus(ctx, spec.TaskID, task.StatusFailed)
		return fmt.Errorf("artifact: %w", err)
	}

	if err := r.tasks.UpdateStatus(ctx, spec.TaskID, task.StatusCompleted); err != nil {
		return err
	}
	trace, _ := r.ids.New(string(ids.PrefixRequest))
	_, _ = r.audit.Record(ctx, auditmod.RecordInput{
		TenantID:  r.tenant,
		TaskID:    spec.TaskID,
		EventType: "task.completed",
		Action:    "task.complete",
		AgentID:   agentID,
		TraceID:   trace,
		Payload:   map[string]any{"profile": manifest.Profile, "agentId": agentID},
	})
	return nil
}

var _ port.AgentRuntime = (*MockRuntime)(nil)
