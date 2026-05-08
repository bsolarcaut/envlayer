package resolver

import "errors"

// ErrNotFound is returned when a key cannot be found in a resolver backend.
var ErrNotFound = errors.New("secret not found")
