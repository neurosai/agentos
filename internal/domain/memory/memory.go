package memory

import "time"

// Type classifies a memory record.
type Type string

const (
	TypeTask        Type = "task"
	TypeSession     Type = "session"
	TypeWorkspace   Type = "workspace"
	TypeCatalogFact Type = "catalog_fact"
	TypeUserPref    Type = "user_pref"
	TypeEvidence    Type = "evidence"
)

// Provenance traces record origins.
type Provenance struct {
	Sources []string
	Notes   string
}

// ACL governs read/manage access by principal references.
type ACL struct {
	View   []string
	Manage []string
}

// Record is a governed memory entry.
type Record struct {
	ID              string
	TenantID        string
	Namespace       string
	Type            Type
	SubjectRef      string
	Content         string
	ContentJSON     map[string]any
	Classification  string
	Confidence      float64
	SourceType      string
	SourceRef       string
	Provenance      Provenance
	ACL             ACL
	RetentionUntil  *time.Time
	CreatedBy       string
	CreatedAt       time.Time
	Supersedes      string
	DeletedAt       *time.Time
	HasEmbedding    bool
}

// Query requests hybrid search over memory.
type Query struct {
	Namespace  string  `json:"namespace,omitempty"`
	QueryText  string  `json:"query,omitempty"`
	Types      []Type  `json:"types,omitempty"`
	Limit      int     `json:"limit,omitempty"`
	MinScore   float64 `json:"minScore,omitempty"`
	SubjectRef string  `json:"subjectRef,omitempty"`
}

// QueryResult is a ranked memory hit.
type QueryResult struct {
	Record Record
	Score  float64
}

// DefaultRetention returns the spec default TTL for a memory type.
func DefaultRetention(t Type) time.Duration {
	switch t {
	case TypeTask:
		return 30 * 24 * time.Hour
	case TypeSession:
		return 7 * 24 * time.Hour
	case TypeEvidence:
		return 90 * 24 * time.Hour
	default:
		return 0
	}
}
