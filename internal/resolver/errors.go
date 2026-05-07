package resolver

import "errors"

// ErrNotFound is returned when a secret key cannot be resolved by a backend.
var ErrNotFound = errors.New("secret not found")

// ErrInvalidConfig is returned when a resolver is misconfigured.
var ErrInvalidConfig = errors.New("invalid resolver configuration")
