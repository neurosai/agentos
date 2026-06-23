package builtin

import (
	"context"
	"fmt"

	"github.com/neurosai/agentos/internal/domain/tool"
)

// Handler executes a builtin tool.
type Handler func(ctx context.Context, args map[string]any) (map[string]any, error)

// Registry maps tool IDs to handlers.
type Registry struct {
	handlers map[string]Handler
}

func NewRegistry() *Registry {
	r := &Registry{handlers: make(map[string]Handler)}
	r.Register("tool.echo", echo)
	r.Register("tool.mock_repo_read", mockRepoRead)
	r.Register("tool.mock_jira_preview", mockJiraPreview)
	r.Register("tool.mock_jira_create", mockJiraCreate)
	return r
}

func (r *Registry) Register(id string, h Handler) {
	r.handlers[id] = h
}

func (r *Registry) Invoke(ctx context.Context, toolID string, args map[string]any) (map[string]any, error) {
	h, ok := r.handlers[toolID]
	if !ok {
		return nil, fmt.Errorf("unknown builtin tool: %s", toolID)
	}
	return h(ctx, args)
}

func echo(_ context.Context, args map[string]any) (map[string]any, error) {
	msg, _ := args["message"].(string)
	return map[string]any{"echo": msg, "args": args}, nil
}

func mockRepoRead(_ context.Context, args map[string]any) (map[string]any, error) {
	path, _ := args["path"].(string)
	return map[string]any{
		"path":    path,
		"content": "// mock file content",
	}, nil
}

func mockJiraPreview(_ context.Context, args map[string]any) (map[string]any, error) {
	title, _ := args["title"].(string)
	return map[string]any{
		"preview": true,
		"title":   title,
		"key":     "MOCK-1",
	}, nil
}

func mockJiraCreate(_ context.Context, args map[string]any) (map[string]any, error) {
	title, _ := args["title"].(string)
	return map[string]any{
		"created": true,
		"key":     "MOCK-42",
		"title":   title,
	}, nil
}

// RefsFromYAML loads tool references from parsed YAML structure.
func RefsFromYAML(entries []tool.Ref) []tool.Ref {
	return entries
}
