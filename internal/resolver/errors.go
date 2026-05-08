package resolver

import "errors"

// ErrNotFound is returned when a resolver cannot locate the requested key.
var ErrNotFound = errors.New("resolver: key not found")
