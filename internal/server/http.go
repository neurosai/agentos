package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/neurosai/agentos/internal/config"
	"github.com/neurosai/agentos/internal/domain/memory"
	"github.com/neurosai/agentos/internal/domain/task"
	"github.com/neurosai/agentos/internal/domain/tool"
	"github.com/neurosai/agentos/internal/domain/agent"
	auditmod "github.com/neurosai/agentos/internal/module/audit"
	memorymod "github.com/neurosai/agentos/internal/module/memory"
	taskmod "github.com/neurosai/agentos/internal/module/task"
	toolmod "github.com/neurosai/agentos/internal/module/tool"
	"github.com/neurosai/agentos/internal/port"
	"github.com/neurosai/agentos/internal/usecase"
	"github.com/neurosai/agentos/pkg/version"
	apperrors "github.com/neurosai/agentos/pkg/errors"
)

// App holds wired modules and dependencies.
type App struct {
	Config  config.Config
	Tasks   *taskmod.Service
	Audit   *auditmod.Service
	Tools   *toolmod.Service
	Memory  *memorymod.Service
	Agent   port.AgentRuntime
	Ready   func(context.Context) error
}

// NewHTTPServer returns the control plane HTTP handler.
func NewHTTPServer(app *App) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		if app.Ready != nil {
			if err := app.Ready(r.Context()); err != nil {
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
				return
			}
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
	})
	mux.HandleFunc("GET /version", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{
			"name":    version.Name,
			"version": version.Version,
		})
	})

	mux.HandleFunc("POST /v1/tasks", app.handleCreateTask)
	mux.HandleFunc("GET /v1/tasks/", app.routeTasks)
	mux.HandleFunc("POST /v1/tasks/", app.routeTasks)

	mux.HandleFunc("GET /v1/audit/events", app.handleQueryAudit)
	mux.HandleFunc("GET /v1/audit/trace/", app.handleAuditTracePrefix)

	mux.HandleFunc("GET /v1/tools", app.handleListTools)
	mux.HandleFunc("GET /v1/tools/", app.routeTools)
	mux.HandleFunc("POST /v1/tools/", app.routeTools)

	mux.HandleFunc("POST /v1/memory/records", app.handleCreateMemory)
	mux.HandleFunc("GET /v1/memory/records/", app.routeMemory)
	mux.HandleFunc("POST /v1/memory/", app.routeMemory)

	return authMiddleware(app.Config)(mux)
}

func (a *App) routeTasks(w http.ResponseWriter, r *http.Request) {
	rest := strings.TrimPrefix(r.URL.Path, "/v1/tasks/")
	if rest == "" {
		http.NotFound(w, r)
		return
	}
	switch {
	case r.Method == http.MethodGet && strings.HasSuffix(rest, "/events"):
		r.SetPathValue("id", strings.TrimSuffix(rest, "/events"))
		a.handleTaskEvents(w, r)
	case r.Method == http.MethodGet && strings.HasSuffix(rest, "/artifacts"):
		r.SetPathValue("id", strings.TrimSuffix(rest, "/artifacts"))
		a.handleListArtifacts(w, r)
	case r.Method == http.MethodPost && strings.HasSuffix(rest, "/approvals"):
		r.SetPathValue("id", strings.TrimSuffix(rest, "/approvals"))
		a.handleDecideApproval(w, r)
	case r.Method == http.MethodPost && strings.HasSuffix(rest, ":cancel"):
		r.SetPathValue("id", strings.TrimSuffix(rest, ":cancel"))
		a.handleCancelTask(w, r)
	case r.Method == http.MethodPost && strings.HasSuffix(rest, ":run"):
		r.SetPathValue("id", strings.TrimSuffix(rest, ":run"))
		a.handleRunTask(w, r)
	case r.Method == http.MethodGet:
		r.SetPathValue("id", rest)
		a.handleGetTask(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (a *App) routeTools(w http.ResponseWriter, r *http.Request) {
	rest := strings.TrimPrefix(r.URL.Path, "/v1/tools/")
	switch {
	case r.Method == http.MethodPost && strings.HasSuffix(rest, ":invoke"):
		r.SetPathValue("id", strings.TrimSuffix(rest, ":invoke"))
		a.handleInvokeTool(w, r)
	case r.Method == http.MethodGet:
		r.SetPathValue("id", rest)
		a.handleGetTool(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (a *App) routeMemory(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/v1/memory/query":
		a.handleQueryMemory(w, r)
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/v1/memory/records/"):
		id := strings.TrimPrefix(r.URL.Path, "/v1/memory/records/")
		r.SetPathValue("id", id)
		a.handleGetMemory(w, r)
	case r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/v1/memory/records/") && strings.HasSuffix(r.URL.Path, ":forget"):
		id := strings.TrimPrefix(r.URL.Path, "/v1/memory/records/")
		id = strings.TrimSuffix(id, ":forget")
		r.SetPathValue("id", id)
		a.handleForgetMemory(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (a *App) handleAuditTracePrefix(w http.ResponseWriter, r *http.Request) {
	traceID := strings.TrimPrefix(r.URL.Path, "/v1/audit/trace/")
	r.SetPathValue("traceId", traceID)
	a.handleAuditTrace(w, r)
}

func authMiddleware(cfg config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/healthz" || r.URL.Path == "/version" {
				next.ServeHTTP(w, r)
				return
			}
			auth := r.Header.Get("Authorization")
			if !strings.HasPrefix(auth, "Bearer ") {
				writeError(w, apperrors.New(apperrors.CodeUnauthenticated, "missing bearer token"))
				return
			}
			token := strings.TrimPrefix(auth, "Bearer ")
			if token != cfg.Dev.BearerToken {
				writeError(w, apperrors.New(apperrors.CodeUnauthenticated, "invalid bearer token"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (a *App) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AgentRef  string            `json:"agentRef"`
		ContextID string            `json:"contextId"`
		Input     map[string]any    `json:"input"`
		Labels    map[string]string `json:"labels"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, apperrors.New(apperrors.CodeInvalidRequest, "invalid json"))
		return
	}
	t, err := a.Tasks.Create(r.Context(), usecase.CreateTaskInput{
		AgentRef:  req.AgentRef,
		ContextID: req.ContextID,
		Input:     req.Input,
		Labels:    req.Labels,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, taskToJSON(t))
}

func (a *App) handleGetTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	t, err := a.Tasks.Get(r.Context(), id)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, taskToJSON(t))
}

func (a *App) handleCancelTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	t, err := a.Tasks.Cancel(r.Context(), id)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, taskToJSON(t))
}

func (a *App) handleRunTask(w http.ResponseWriter, r *http.Request) {
	if a.Agent == nil {
		writeError(w, apperrors.New(apperrors.CodeBackendUnavailable, "agent runtime not configured"))
		return
	}
	id := r.PathValue("id")
	var req struct {
		AgentRef     string `json:"agentRef"`
		ManifestPath string `json:"manifestPath"`
	}
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}
	t, err := a.Tasks.Get(r.Context(), id)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	agentRef := req.AgentRef
	if agentRef == "" {
		agentRef = t.AgentRef
	}
	spec := agent.RunSpec{
		TaskID:       id,
		AgentRef:     agentRef,
		ManifestPath: req.ManifestPath,
	}
	go func() {
		ctx := context.Background()
		if err := a.Agent.Run(ctx, spec); err != nil {
			_, _ = fmt.Fprintf(io.Discard, "agent run %s: %v\n", id, err)
		}
	}()
	writeJSON(w, http.StatusAccepted, map[string]any{
		"id":     id,
		"status": "RUNNING",
	})
}

func (a *App) handleTaskEvents(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	ch, err := a.Tasks.StreamEvents(r.Context(), id, true)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	for msg := range ch {
		fmt.Fprintf(w, "data: %s\n\n", msg.Content)
		flusher.Flush()
	}
}

func (a *App) handleDecideApproval(w http.ResponseWriter, r *http.Request) {
	taskID := r.PathValue("id")
	var req struct {
		Approved bool   `json:"approved"`
		Reason   string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, apperrors.New(apperrors.CodeInvalidRequest, "invalid json"))
		return
	}
	t, err := a.Tasks.DecideApproval(r.Context(), taskID, "", req.Approved, a.Config.Dev.SubjectID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	if req.Approved && a.Tools != nil {
		_, _ = a.Tools.ResumeAfterApproval(r.Context(), taskID)
	}
	writeJSON(w, http.StatusOK, taskToJSON(t))
}

func (a *App) handleListArtifacts(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	arts, err := a.Tasks.ListArtifacts(r.Context(), id)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, arts)
}

func (a *App) handleQueryAudit(w http.ResponseWriter, r *http.Request) {
	events, err := a.Audit.Query(r.Context(), r.URL.Query().Get("tenantId"), r.URL.Query().Get("taskId"), 100)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, events)
}

func (a *App) handleAuditTrace(w http.ResponseWriter, r *http.Request) {
	traceID := r.PathValue("traceId")
	events, err := a.Audit.QueryByTrace(r.Context(), traceID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, events)
}

func (a *App) handleListTools(w http.ResponseWriter, r *http.Request) {
	tools, err := a.Tools.List(r.Context(), a.Config.Dev.TenantID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, tools)
}

func (a *App) handleGetTool(w http.ResponseWriter, r *http.Request) {
	ref, err := a.Tools.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ref)
}

func (a *App) handleInvokeTool(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TaskID         string         `json:"taskId"`
		AgentID        string         `json:"agentId"`
		Arguments      map[string]any `json:"arguments"`
		IdempotencyKey string         `json:"idempotencyKey"`
		Context        map[string]any `json:"context"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, apperrors.New(apperrors.CodeInvalidRequest, "invalid json"))
		return
	}
	invoke := tool.Invoke{
		TaskID:         req.TaskID,
		AgentID:        req.AgentID,
		ToolID:         r.PathValue("id"),
		Arguments:      req.Arguments,
		IdempotencyKey: req.IdempotencyKey,
	}
	if req.Context != nil {
		if v, ok := req.Context["classification"].(string); ok {
			invoke.Context.Classification = v
		}
		if v, ok := req.Context["sourceTrust"].(string); ok {
			invoke.Context.SourceTrust = v
		}
	}
	res, err := a.Tools.Invoke(r.Context(), invoke)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (a *App) handleCreateMemory(w http.ResponseWriter, r *http.Request) {
	var req memory.Record
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, apperrors.New(apperrors.CodeInvalidRequest, "invalid json"))
		return
	}
	rec, err := a.Memory.Create(r.Context(), req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, rec)
}

func (a *App) handleGetMemory(w http.ResponseWriter, r *http.Request) {
	rec, err := a.Memory.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rec)
}

func (a *App) handleQueryMemory(w http.ResponseWriter, r *http.Request) {
	var q memory.Query
	if err := json.NewDecoder(r.Body).Decode(&q); err != nil {
		writeError(w, apperrors.New(apperrors.CodeInvalidRequest, "invalid json"))
		return
	}
	results, err := a.Memory.Search(r.Context(), q)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, results)
}

func (a *App) handleForgetMemory(w http.ResponseWriter, r *http.Request) {
	if err := a.Memory.Delete(r.Context(), r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func taskToJSON(t task.Task) map[string]any {
	return map[string]any{
		"id":        t.ID,
		"tenantId":  t.TenantID,
		"contextId": t.ContextID,
		"agentRef":  t.AgentRef,
		"status":    t.Status,
		"input":     t.Input,
		"labels":    t.Labels,
		"createdBy": t.CreatedBy,
		"createdAt": t.CreatedAt.Format(time.RFC3339),
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, err *apperrors.Error) {
	writeJSON(w, err.StatusCode, map[string]any{
		"code":    err.Code,
		"message": err.Message,
	})
}

func writeServiceError(w http.ResponseWriter, err error) {
	if policymodDenied(err) {
		writeError(w, apperrors.New(apperrors.CodePolicyDenied, err.Error()))
		return
	}
	if strings.Contains(err.Error(), "not found") {
		writeError(w, apperrors.New(apperrors.CodeNotFound, err.Error()))
		return
	}
	writeError(w, apperrors.New(apperrors.CodeBackendUnavailable, err.Error()))
}

func policymodDenied(err error) bool {
	return strings.Contains(err.Error(), "policy denied")
}

// DrainBody closes request body.
func DrainBody(r *http.Request) {
	if r.Body != nil {
		_, _ = io.Copy(io.Discard, r.Body)
		_ = r.Body.Close()
	}
}
