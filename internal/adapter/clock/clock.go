package clock

import "time"

// Real is a production Clock implementation.
type Real struct{}

func (Real) Now() time.Time { return time.Now() }
