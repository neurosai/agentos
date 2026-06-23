package idgen

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/neurosai/agentos/internal/port"
	"github.com/neurosai/agentos/pkg/ids"
)

// UUID generates opaque ID suffixes for prefixed identifiers.
type UUID struct{}

func (UUID) New(prefix string) (string, error) {
	var b [12]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	p, err := ids.New(ids.Prefix(prefix), hex.EncodeToString(b[:]))
	if err != nil {
		return "", err
	}
	return p.String(), nil
}

// MustNew panics on failure.
func MustNew(prefix string) string {
	id, err := UUID{}.New(prefix)
	if err != nil {
		panic(err)
	}
	return id
}

var _ port.IDGenerator = UUID{}

// NewID creates a prefixed ID using the default generator.
func NewID(prefix ids.Prefix) (string, error) {
	return UUID{}.New(string(prefix))
}

// Format validates prefix name.
func Format(prefix ids.Prefix, value string) (string, error) {
	id, err := ids.New(prefix, value)
	if err != nil {
		return "", fmt.Errorf("new id: %w", err)
	}
	return id.String(), nil
}
