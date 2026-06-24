package agentmod

import (
	"fmt"
	"os"

	"github.com/neurosai/agentos/internal/domain/agent"
	"gopkg.in/yaml.v3"
)

type manifestYAML struct {
	Metadata struct {
		Name   string `yaml:"name"`
		Labels map[string]string `yaml:"labels"`
	} `yaml:"metadata"`
	Spec struct {
		Identity struct {
			Principal string `yaml:"principal"`
		} `yaml:"identity"`
	} `yaml:"spec"`
}

// LoadManifest reads an agent manifest YAML file.
func LoadManifest(path string) (agent.Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return agent.Manifest{}, fmt.Errorf("read manifest: %w", err)
	}
	var doc manifestYAML
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return agent.Manifest{}, fmt.Errorf("parse manifest: %w", err)
	}
	m := agent.Manifest{
		Name:    doc.Metadata.Name,
		Profile: doc.Metadata.Labels["profile"],
		AgentID: doc.Spec.Identity.Principal,
	}
	if m.AgentID == "" {
		m.AgentID = "agent:" + doc.Metadata.Name
	}
	return m, nil
}
