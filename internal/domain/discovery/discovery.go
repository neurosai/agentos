package discovery

import "time"

// CollectorKind identifies a safe discovery collector.
type CollectorKind string

const (
	CollectorGit       CollectorKind = "git"
	CollectorKubernetes CollectorKind = "kubernetes"
	CollectorAPI       CollectorKind = "api_descriptor"
	CollectorCICD      CollectorKind = "cicd"
	CollectorOTel      CollectorKind = "otel_metadata"
)

// Mode restricts collector behavior.
type Mode string

const (
	ModeReadOnly Mode = "read_only"
)

// JobStatus tracks discovery job lifecycle.
type JobStatus string

const (
	JobPending   JobStatus = "pending"
	JobRunning   JobStatus = "running"
	JobCompleted JobStatus = "completed"
	JobFailed    JobStatus = "failed"
	JobCancelled JobStatus = "cancelled"
)

// Scope bounds what a collector may read.
type Scope struct {
	Cluster    string
	Namespaces []string
	Repo       string
	Paths      []string
}

// Job requests a safe collector run.
type Job struct {
	ID             string
	TenantID       string
	Collector      CollectorKind
	Scope          Scope
	Mode           Mode
	WriteToCatalog bool
	WriteToMemory  bool
	RequestedBy    string
	Status         JobStatus
	CreatedAt      time.Time
	CompletedAt    *time.Time
}

// Observation is a single collected fact.
type Observation struct {
	ID             string
	JobID          string
	Collector      CollectorKind
	ObservedAt     time.Time
	ResourceRef    string
	Kind           string
	Claim          map[string]any
	Evidence       map[string]any
	Classification string
	Confidence     float64
}

// UnsafeCollector reports collectors forbidden in v0.1.
func UnsafeCollector(name string) bool {
	switch name {
	case "packet_capture", "host_scan", "credential_guess", "secret_read", "network_sniff":
		return true
	default:
		return false
	}
}

// ValidateJob ensures the job uses only safe collectors and read-only mode.
func ValidateJob(j Job) error {
	if UnsafeCollector(string(j.Collector)) {
		return ErrUnsafeCollector
	}
	if j.Mode != ModeReadOnly {
		return ErrInvalidMode
	}
	if j.WriteToMemory && !j.WriteToCatalog {
		return ErrMemoryWithoutCatalog
	}
	return nil
}
