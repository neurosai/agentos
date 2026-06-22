package ids

import (
	"fmt"
	"strings"
)

// Prefix identifies the entity type encoded in a prefixed ID.
type Prefix string

const (
	PrefixTask     Prefix = "task"
	PrefixEvent    Prefix = "evt"
	PrefixMemory   Prefix = "mem"
	PrefixObs      Prefix = "obs"
	PrefixDecision Prefix = "dec"
	PrefixToolCall Prefix = "toolcall"
	PrefixRequest  Prefix = "req"
	PrefixAgentRun Prefix = "run"
)

// ID is a typed prefixed identifier (e.g. task_01JY4F...).
type ID struct {
	Prefix Prefix
	Value  string
}

func (id ID) String() string {
	if id.Value == "" {
		return ""
	}
	return string(id.Prefix) + "_" + id.Value
}

// Parse splits a prefixed ID into prefix and value.
func Parse(raw string) (ID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ID{}, fmt.Errorf("empty id")
	}
	idx := strings.Index(raw, "_")
	if idx <= 0 || idx >= len(raw)-1 {
		return ID{}, fmt.Errorf("invalid id format: %q", raw)
	}
	return ID{Prefix: Prefix(raw[:idx]), Value: raw[idx+1:]}, nil
}

// MustParse parses raw or panics. Intended for tests.
func MustParse(raw string) ID {
	id, err := Parse(raw)
	if err != nil {
		panic(err)
	}
	return id
}

// Validate checks that raw matches the expected prefix.
func Validate(raw string, want Prefix) error {
	id, err := Parse(raw)
	if err != nil {
		return err
	}
	if id.Prefix != want {
		return fmt.Errorf("expected prefix %q, got %q", want, id.Prefix)
	}
	if len(id.Value) < 8 {
		return fmt.Errorf("id value too short")
	}
	return nil
}

// New constructs a prefixed ID from prefix and opaque value.
func New(prefix Prefix, value string) (ID, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return ID{}, fmt.Errorf("empty id value")
	}
	id := ID{Prefix: prefix, Value: value}
	if err := Validate(id.String(), prefix); err != nil {
		return ID{}, err
	}
	return id, nil
}

// TaskID is a task identifier.
type TaskID = ID

// EventID is an audit or task event identifier.
type EventID = ID

// MemoryID is a memory record identifier.
type MemoryID = ID

// ObservationID is a discovery observation identifier.
type ObservationID = ID

// DecisionID is a policy decision identifier.
type DecisionID = ID
