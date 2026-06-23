package builtin

import (
	"context"
	"fmt"
	"os"

	"github.com/neurosai/agentos/internal/domain/tool"
	"gopkg.in/yaml.v3"
)

type toolsFile struct {
	Tools []toolRefYAML `yaml:"tools"`
}

type toolRefYAML struct {
	ID          string `yaml:"id"`
	Name        string `yaml:"name"`
	Transport   string `yaml:"transport"`
	Risk        string `yaml:"risk"`
	Description string `yaml:"description"`
}

// LoadRegistry reads deploy/tools.yaml into tool.Ref slice and registry.
func LoadRegistry(path string) ([]tool.Ref, *Registry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	var f toolsFile
	if err := yaml.Unmarshal(data, &f); err != nil {
		return nil, nil, err
	}
	refs := make([]tool.Ref, 0, len(f.Tools))
	for _, t := range f.Tools {
		refs = append(refs, tool.Ref{
			ID:          t.ID,
			Name:        t.Name,
			Transport:   t.Transport,
			Risk:        t.Risk,
			Description: t.Description,
		})
	}
	reg := NewRegistry()
	return refs, reg, nil
}

// StaticRegistry implements ToolRegistry from a slice.
type StaticRegistry struct {
	refs map[string]tool.Ref
}

func NewStaticRegistry(refs []tool.Ref) *StaticRegistry {
	m := make(map[string]tool.Ref, len(refs))
	for _, r := range refs {
		m[r.ID] = r
	}
	return &StaticRegistry{refs: m}
}

func (s *StaticRegistry) List(ctx context.Context, tenantID string) ([]tool.Ref, error) {
	out := make([]tool.Ref, 0, len(s.refs))
	for _, r := range s.refs {
		out = append(out, r)
	}
	return out, nil
}

func (s *StaticRegistry) Get(ctx context.Context, toolID string) (tool.Ref, error) {
	r, ok := s.refs[toolID]
	if !ok {
		return tool.Ref{}, fmt.Errorf("tool not found: %s", toolID)
	}
	return r, nil
}

func (s *StaticRegistry) Upsert(ctx context.Context, ref tool.Ref) error {
	s.refs[ref.ID] = ref
	return nil
}
