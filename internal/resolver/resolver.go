package resolver

import (
	"context"
	"errors"
)

// ErrNotFound is returned by a Resolver when the requested key does not exist
// in the backing store.
var ErrNotFound = errors.New("resolver: key not found")

// Resolver is the interface that all secret backends must implement.
type Resolver interface {
	// Name returns a human-readable identifier for the resolver (e.g. "vault", "ssm", "dotenv").
	Name() string

	// Resolve returns the plaintext value associated with key, or ErrNotFound if
	// the key is not present in this backend. Any other error signals a backend
	// failure.
	Resolve(ctx context.Context, key string) (string, error)
}

// Chain combines multiple Resolvers into a single Resolver that tries each
// backend in order, returning the first successful result.
func NewChain(resolvers ...Resolver) Resolver {
	return &chain{resolvers: resolvers}
}
