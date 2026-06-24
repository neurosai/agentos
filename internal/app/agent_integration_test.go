//go:build integration

package app_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/neurosai/agentos/internal/app"
	"github.com/neurosai/agentos/internal/adapter/postgres"
	"github.com/neurosai/agentos/internal/config"
	"github.com/neurosai/agentos/internal/domain/agent"
	"github.com/neurosai/agentos/internal/domain/task"
	"github.com/neurosai/agentos/internal/usecase"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func startPostgres(t *testing.T) (string, func()) {
	t.Helper()
	if os.Getenv("SKIP_TESTCONTAINERS") == "1" {
		t.Skip("SKIP_TESTCONTAINERS=1")
	}
	ctx := context.Background()
	container, err := postgres.Run(ctx,
		"pgvector/pgvector:pg17",
		postgres.WithDatabase("agentos"),
		postgres.WithUsername("agentos"),
		postgres.WithPassword("agentos"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
		),
	)
	require.NoError(t, err)
	url, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)
	cleanup := func() { _ = container.Terminate(ctx) }
	return url, cleanup
}

func TestAgentSubmitIntegration(t *testing.T) {
	url, cleanup := startPostgres(t)
	defer cleanup()
	ctx := context.Background()

	root := filepath.Join("..", "..")
	cfg := config.Config{
		Listen: ":0",
		Database: config.DatabaseConfig{
			URL:           url,
			AutoMigrate:   true,
		},
		Dev: config.DevConfig{
			TenantID:      "tenant:dev",
			SubjectID:     "user:dev",
			SubjectRoles:  []string{"engineer"},
			SubjectGroups: []string{"group:team-payments", "group:team-platform"},
			BearerToken:   "dev-token",
		},
		Tools: config.ToolsConfig{Registry: filepath.Join(root, "deploy/tools.yaml")},
	}

	srvApp, cleanupApp, err := app.Build(ctx, cfg)
	require.NoError(t, err)
	defer cleanupApp()

	created, err := srvApp.Tasks.Create(ctx, usecase.CreateTaskInput{
		AgentRef: "agent:hermes-dev",
		Input:    map[string]any{"goal": "Analyze payments repository"},
		Labels:   map[string]string{"classification": "internal"},
	})
	require.NoError(t, err)

	manifest := filepath.Join(root, "examples", "agents", "hermes-dev.yaml")
	require.NoError(t, srvApp.Agent.Run(ctx, agent.RunSpec{
		TaskID:       created.ID,
		AgentRef:     "agent:hermes-dev",
		ManifestPath: manifest,
	}))

	got, err := srvApp.Tasks.Get(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, task.StatusCompleted, got.Status)

	arts, err := srvApp.Tasks.ListArtifacts(ctx, created.ID)
	require.NoError(t, err)
	require.Len(t, arts, 1)

	events, err := srvApp.Audit.Query(ctx, created.TenantID, created.ID, 50)
	require.NoError(t, err)
	require.NotEmpty(t, events)
}
