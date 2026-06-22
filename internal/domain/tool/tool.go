package tool

import "time"

// CredentialMode describes how ToolD resolves downstream credentials.
type CredentialMode string

const (
	CredentialNone     CredentialMode = "none"
	CredentialExchange CredentialMode = "exchange"
	CredentialVaultRef CredentialMode = "vault_ref"
)

// Ref identifies a registered tool.
type Ref struct {
	ID          string
	Name        string
	Transport   string
	MCPServer   string
	Risk        string
	Description string
}

// InvokeContext carries trust and classification metadata for a call.
type InvokeContext struct {
	Classification string
	SourceTrust    string
	Workspace      string
}

// Invoke is a policy-bound tool invocation request.
type Invoke struct {
	TaskID         string
	AgentID        string
	ToolID         string
	Arguments      map[string]any
	CredentialMode CredentialMode
	Context        InvokeContext
}

// Result captures tool output and audit correlation.
type Result struct {
	CallID       string
	Status       string
	ContentType  string
	Output       map[string]any
	OutputTrust  string
	AuditEventID string
	CompletedAt  time.Time
}

// InvocationRecord is the persisted audit cross-reference.
type InvocationRecord struct {
	ID           string
	ToolID       string
	TaskID       string
	AgentID      string
	Status       string
	AuditEventID string
	StartedAt    time.Time
	CompletedAt  *time.Time
}
