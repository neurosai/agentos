package agent

// Manifest is a minimal agent profile for mock runtime v0.3.
type Manifest struct {
	Name    string
	Profile string
	AgentID string
}

// RunSpec links a task to an agent execution.
type RunSpec struct {
	TaskID       string
	AgentRef     string
	ManifestPath string
	Manifest     Manifest
}
