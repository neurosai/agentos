//go:build integration

package migrations_test

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func migrationsDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Clean(filepath.Join(filepath.Dir(file), "../../migrations"))
}

func TestMigrationsUpDown(t *testing.T) {
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
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(60*time.Second),
		),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, container.Terminate(ctx))
	})

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := sql.Open("pgx", connStr)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	dir := migrationsDir(t)
	require.NoError(t, goose.SetDialect("postgres"))

	require.NoError(t, goose.Up(db, dir))

	var tableCount int
	require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE'`).Scan(&tableCount))
	require.GreaterOrEqual(t, tableCount, 10)

	require.NoError(t, goose.DownTo(db, dir, 0))
}
