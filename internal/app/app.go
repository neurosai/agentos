package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/neurosai/agentos/internal/adapter/builtin"
	"github.com/neurosai/agentos/internal/adapter/clock"
	"github.com/neurosai/agentos/internal/adapter/idgen"
	memadapter "github.com/neurosai/agentos/internal/adapter/memory"
	"github.com/neurosai/agentos/internal/adapter/opa"
	"github.com/neurosai/agentos/internal/adapter/postgres"
	"github.com/neurosai/agentos/internal/config"
	auditmod "github.com/neurosai/agentos/internal/module/audit"
	memorymod "github.com/neurosai/agentos/internal/module/memory"
	policymod "github.com/neurosai/agentos/internal/module/policy"
	taskmod "github.com/neurosai/agentos/internal/module/task"
	toolmod "github.com/neurosai/agentos/internal/module/tool"
	"github.com/neurosai/agentos/internal/port"
	"github.com/neurosai/agentos/internal/server"
)

// Build wires dependencies and returns a runnable App.
func Build(ctx context.Context, cfg config.Config) (*server.App, func(), error) {
	clk := clock.Real{}
	ids := idgen.UUID{}
	broadcast := memadapter.NewEventBroadcaster()

	var taskRepo port.TaskRepository = memadapter.NewTaskRepository()
	var msgRepo port.TaskMessageRepository = memadapter.NewTaskMessageRepository()
	var artRepo port.TaskArtifactRepository = memadapter.NewTaskArtifactRepository()
	var apprRepo port.TaskApprovalRepository = memadapter.NewTaskApprovalRepository()
	var auditRepo port.AuditRepository = memadapter.NewAuditRepository()
	var memRepo port.MemoryRepository = memadapter.NewMemoryRepository()
	var invRepo port.ToolInvocationRepository = memadapter.NewToolInvocationRepository()

	var pool *postgres.Pool
	var cleanup func() = func() {}

	if cfg.Database.URL != "" {
		if cfg.Database.AutoMigrate {
			if err := postgres.Migrate(ctx, cfg.Database.URL); err != nil {
				return nil, nil, fmt.Errorf("migrate: %w", err)
			}
		}
		p, err := postgres.NewPool(ctx, cfg.Database.URL)
		if err != nil {
			return nil, nil, err
		}
		pool = p
		cleanup = func() { pool.Close() }
		taskRepo = postgres.NewTaskRepository(pool)
		msgRepo = postgres.NewTaskMessageRepository(pool)
		artRepo = postgres.NewTaskArtifactRepository(pool)
		apprRepo = postgres.NewTaskApprovalRepository(pool)
		auditRepo = postgres.NewAuditRepository(pool)
		memRepo = postgres.NewMemoryRepository(pool)
		invRepo = postgres.NewToolInvocationRepository(pool)
	}

	auditSvc := auditmod.NewService(auditRepo, clk, ids)
	var evaluator opa.Evaluator
	if cfg.OPA.URL != "" {
		evaluator = opa.NewClient(cfg.OPA.URL)
	} else {
		evaluator = opa.RequireHighRisk{}
	}
	policySvc := policymod.NewService(evaluator, auditSvc, cfg.Dev.TenantID, cfg.Dev.SubjectID)

	taskSvc := taskmod.NewService(taskmod.Options{
		Tasks:     taskRepo,
		Messages:  msgRepo,
		Artifacts: artRepo,
		Approvals: apprRepo,
		Policy:    policySvc,
		Audit:     auditSvc,
		Clock:     clk,
		IDs:       ids,
		Tenant:    cfg.Dev.TenantID,
		Subject:   cfg.Dev.SubjectID,
		Roles:     cfg.Dev.SubjectRoles,
		Broadcast: broadcast,
	})

	toolPath := cfg.Tools.Registry
	if toolPath == "" {
		toolPath = "deploy/tools.yaml"
	}
	refs, invoker, err := builtin.LoadRegistry(toolPath)
	if err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("load tools: %w", err)
	}
	reg := builtin.NewStaticRegistry(refs)

	toolSvc := toolmod.NewService(toolmod.Options{
		Registry:    reg,
		Invoker:     invoker,
		Invocations: invRepo,
		Policy:      policySvc,
		Audit:       auditSvc,
		Tasks:       taskSvc,
		Clock:       clk,
		IDs:         ids,
		Tenant:      cfg.Dev.TenantID,
		Subject:     cfg.Dev.SubjectID,
		Roles:       cfg.Dev.SubjectRoles,
	})

	memSvc := memorymod.NewService(memorymod.Options{
		Repo:    memRepo,
		Policy:  policySvc,
		Audit:   auditSvc,
		Clock:   clk,
		IDs:     ids,
		Tenant:  cfg.Dev.TenantID,
		Subject: cfg.Dev.SubjectID,
		Roles:   cfg.Dev.SubjectRoles,
	})

	ready := func(ctx context.Context) error {
		if pool != nil {
			if err := pool.Ping(ctx); err != nil {
				return err
			}
		}
		if cfg.OPA.URL != "" {
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.OPA.URL+"/health", nil)
			if err != nil {
				return err
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			_ = resp.Body.Close()
			if resp.StatusCode >= 300 {
				return fmt.Errorf("opa not ready")
			}
		}
		return nil
	}

	return &server.App{
		Config: cfg,
		Tasks:  taskSvc,
		Audit:  auditSvc,
		Tools:  toolSvc,
		Memory: memSvc,
		Ready:  ready,
	}, cleanup, nil
}

// Serve starts HTTP until ctx cancelled.
func Serve(ctx context.Context, cfg config.Config) error {
	app, cleanup, err := Build(ctx, cfg)
	if err != nil {
		return err
	}
	defer cleanup()

	srv := &http.Server{
		Addr:    cfg.Listen,
		Handler: server.NewHTTPServer(app),
	}
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdown)
	}()
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}
