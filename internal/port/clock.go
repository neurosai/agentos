package port

import "time"

// Clock abstracts time for testability.
type Clock interface {
	Now() time.Time
}

// IDGenerator creates opaque identifier suffixes.
type IDGenerator interface {
	New(prefix string) (string, error)
}
