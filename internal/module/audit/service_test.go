package auditmod_test

import (
	"context"
	"testing"
	"time"

	"github.com/neurosai/agentos/internal/adapter/clock"
	"github.com/neurosai/agentos/internal/adapter/idgen"
	mem "github.com/neurosai/agentos/internal/adapter/memory"
	auditmod "github.com/neurosai/agentos/internal/module/audit"
	"github.com/stretchr/testify/require"
)

func TestAuditHashChain(t *testing.T) {
	repo := mem.NewAuditRepository()
	svc := auditmod.NewService(repo, clock.Real{}, idgen.UUID{})
	ctx := context.Background()

	e1, err := svc.Record(ctx, auditmod.RecordInput{
		TenantID:  "tenant:a",
		EventType: "task.created",
		Payload:   map[string]any{"n": 1},
	})
	require.NoError(t, err)
	require.NotEmpty(t, e1.EventHash)
	require.Empty(t, e1.PrevHash)

	e2, err := svc.Record(ctx, auditmod.RecordInput{
		TenantID:  "tenant:a",
		EventType: "policy.decided",
		Payload:   map[string]any{"n": 2},
	})
	require.NoError(t, err)
	require.Equal(t, e1.EventHash, e2.PrevHash)

	proof, err := svc.Proof(ctx, "tenant:a")
	require.NoError(t, err)
	require.Equal(t, e2.EventHash, proof.RootHash)
}

func TestAuditQueryByTrace(t *testing.T) {
	repo := mem.NewAuditRepository()
	svc := auditmod.NewService(repo, clock.Real{}, idgen.UUID{})
	ctx := context.Background()
	_, _ = svc.Record(ctx, auditmod.RecordInput{TenantID: "t", TraceID: "trace-1", EventType: "a"})
	_, _ = svc.Record(ctx, auditmod.RecordInput{TenantID: "t", TraceID: "trace-1", EventType: "b"})
	events, err := svc.QueryByTrace(ctx, "trace-1")
	require.NoError(t, err)
	require.Len(t, events, 2)
	_ = time.Now()
}
