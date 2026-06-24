package agentmod_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/neurosai/agentos/internal/adapter/builtin"
	"github.com/neurosai/agentos/internal/adapter/clock"
	"github.com/neurosai/agentos/internal/adapter/idgen"
	memadapter "github.com/neurosai/agentos/internal/adapter/memory"
	"github.com/neurosai/agentos/internal/adapter/opa"
	"github.com/neurosai/agentos/internal/domain/agent"
	"github.com/neurosai/agentos/internal/domain/task"
	agentmod "github.com/neurosai/agentos/internal/module/agent"
	auditmod "github.com/neurosai/agentos/internal/module/audit"
	memorymod "github.com/neurosai/agentos/internal/module/memory"
	policymod "github.com/neurosai/agentos/internal/module/policy"
	taskmod "github.com/neurosai/agentos/internal/module/task"
	toolmod "github.com/neurosai/agentos/internal/module/tool"
	"github.com/neurosai/agentos/internal/usecase"
	"github.com/stretchr/testify/require"
)

func wireMockRuntime(t *testing.T) (*agentmod.MockRuntime, *taskmod.Service) {
	t.Helper()
	root := filepath.Join("..", "..", "..")
	clk := clock.Real{}
	ids := idgen.UUID{}
	broadcast := memadapter.NewEventBroadcaster()
	auditSvc := auditmod.NewService(memadapter.NewAuditRepository(), clk, ids)
	policySvc := policymod.NewService(opa.RequireHighRisk{}, auditSvc, "tenant:dev", "user:dev")

	taskSvc := taskmod.NewService(taskmod.Options{
		Tasks:     memadapter.NewTaskRepository(),
		Messages:  memadapter.NewTaskMessageRepository(),
		Artifacts: memadapter.NewTaskArtifactRepository(),
		Approvals: memadapter.NewTaskApprovalRepository(),
		Policy:    policySvc,
		Audit:     auditSvc,
		Clock:     clk,
		IDs:       ids,
		Tenant:    "tenant:dev",
		Subject:   "user:dev",
		Roles:     []string{"engineer"},
		Broadcast: broadcast,
	})

	refs, invoker, err := builtin.LoadRegistry(filepath.Join(root, "deploy/tools.yaml"))
	require.NoError(t, err)
	reg := builtin.NewStaticRegistry(refs)
	toolSvc := toolmod.NewService(toolmod.Options{
		Registry:    reg,
		Invoker:     invoker,
		Invocations: memadapter.NewToolInvocationRepository(),
		Policy:      policySvc,
		Audit:       auditSvc,
		Tasks:       taskSvc,
		Clock:       clk,
		IDs:         ids,
		Tenant:      "tenant:dev",
		Subject:     "user:dev",
		Roles:       []string{"engineer"},
	})
	memSvc := memorymod.NewService(memorymod.Options{
		Repo:    memadapter.NewMemoryRepository(),
		Policy:  policySvc,
		Audit:   auditSvc,
		Clock:   clk,
		IDs:     ids,
		Tenant:  "tenant:dev",
		Subject: "user:dev",
		Roles:   []string{"engineer"},
		Groups:  []string{"group:team-payments"},
	})
	rt := agentmod.NewMockRuntime(agentmod.MockOptions{
		Tasks:  taskSvc,
		Tools:  toolSvc,
		Memory: memSvc,
		Audit:  auditSvc,
		IDs:    ids,
		Tenant: "tenant:dev",
	})
	return rt, taskSvc
}

func TestMockRuntimeCompletesTask(t *testing.T) {
	rt, taskSvc := wireMockRuntime(t)
	ctx := context.Background()
	created, err := taskSvc.Create(ctx, usecase.CreateTaskInput{
		AgentRef: "agent:hermes-dev",
		Input:    map[string]any{"goal": "analyze payments"},
		Labels:   map[string]string{"classification": "internal"},
	})
	require.NoError(t, err)

	manifestPath := filepath.Join("..", "..", "..", "examples", "agents", "hermes-dev.yaml")
	err = rt.Run(ctx, agent.RunSpec{
		TaskID:       created.ID,
		AgentRef:     "agent:hermes-dev",
		ManifestPath: manifestPath,
	})
	require.NoError(t, err)

	got, err := taskSvc.Get(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, task.StatusCompleted, got.Status)

	arts, err := taskSvc.ListArtifacts(ctx, created.ID)
	require.NoError(t, err)
	require.Len(t, arts, 1)
	require.Equal(t, "run-summary", arts[0].Name)
}

func TestLoadManifest(t *testing.T) {
	path := filepath.Join("..", "..", "..", "examples", "agents", "hermes-dev.yaml")
	m, err := agentmod.LoadManifest(path)
	require.NoError(t, err)
	require.Equal(t, "hermes-dev", m.Name)
	require.Equal(t, "repo-analysis", m.Profile)
	require.Equal(t, "agent:hermes-dev", m.AgentID)
}

func TestMockRuntimeOrder(t *testing.T) {
	rt, taskSvc := wireMockRuntime(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	created, err := taskSvc.Create(ctx, usecase.CreateTaskInput{
		AgentRef: "agent:demo",
		Input:    map[string]any{"goal": "demo"},
	})
	require.NoError(t, err)
	require.NoError(t, rt.Run(ctx, agent.RunSpec{TaskID: created.ID, AgentRef: "agent:demo"}))
}
