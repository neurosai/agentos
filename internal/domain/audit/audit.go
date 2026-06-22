package audit

import "time"

// Event is an append-only audit record with hash-chain fields.
type Event struct {
	EventID      string
	OccurredAt   time.Time
	TenantID     string
	SubjectID    string
	AgentID      string
	TaskID       string
	EventType    string
	ResourceType string
	ResourceID   string
	Action       string
	Decision     string
	Status       string
	PayloadHash  string
	PrevHash     string
	EventHash    string
	TraceID      string
	SpanID       string
}

// Proof documents a hash-chain segment for tamper evidence.
type Proof struct {
	StreamID   string
	FromEvent  string
	ToEvent    string
	RootHash   string
	AnchorRef  string
	VerifiedAt time.Time
}

// Anchor records an external digest snapshot.
type Anchor struct {
	StreamID   string
	Digest     string
	AnchoredAt time.Time
	StorageRef string
}

// Valid checks required hash-chain fields are present.
func (e Event) Valid() bool {
	return e.EventID != "" && e.OccurredAt.IsZero() == false &&
		e.EventHash != "" && e.PayloadHash != ""
}
