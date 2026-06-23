//go:build integration

package postgres_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/neurosai/agentos/internal/adapter/postgres"
	"github.com/neurosai/agentos/internal/domain/audit"
	"github.com/neurosai/agentos/internal/domain/task"
	"github.com/neurosai/agentos/internal/port"
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

func TestTaskAndAuditRepos(t *testing.T) {
	url, cleanup := startPostgres(t)
	defer cleanup()
	ctx := context.Background()
	require.NoError(t, postgres.Migrate(ctx, url))
	pool, err := postgres.NewPool(ctx, url)
	require.NoError(t, err)
	defer pool.Close()

	taskRepo := postgres.NewTaskRepository(pool)
	auditRepo := postgres.NewAuditRepository(pool)

	now := time.Now().UTC()
	id := "task_testintegration01"
	require.NoError(t, taskRepo.Create(ctx, task.Task{
		ID: id, TenantID: "tenant:t", AgentRef: "agent:a", Status: task.StatusAccepted,
		Input: map[string]any{"goal": "test"}, Labels: map[string]string{}, CreatedBy: "u",
		CreatedAt: now, UpdatedAt: now,
	}))
	got, err := taskRepo.Get(ctx, id)
	require.NoError(t, err)
	require.Equal(t, task.StatusAccepted, got.Status)

	ev := audit.Event{
		EventID: "evt_testintegration01", OccurredAt: now, TenantID: "tenant:t",
		EventType: "task.created", PayloadHash: "abc", EventHash: "def",
	}
	require.NoError(t, auditRepo.Insert(ctx, ev))
	events, err := auditRepo.List(ctx, port.AuditQueryFilter{TenantID: "tenant:t"})
	require.NoError(t, err)
	require.Len(t, events, 1)
}
